# Apptainer

Apptainer是一款面向高性能计算领域的容器软件。本分支基于Apptainer 1.0.0版本，给它增加了检查点的支持。

## 检查点集成

检查点功能基于CRIU 3.18以上版本（因为依赖non-root特性）。

### 管理检查点目录相关命令

- `apptainer checkpoint create [--mem-dir] <name>` 创建检查点目录

会在用户的home目录下的默认路径`~/.apptainer/checkpoint/criu/<name>`创建一个目录，用于存储相关信息。这个目录是一个广义的检查点目录，容器生成检查点时，会将容器配置导出到这个目录下，CRIU工作时也会将输出日志写入这个文件，容器运行时也会将工作进程的进程号写入这个目录下，以供CRIU读取。

另外，容器检查点目录的子目录`~/.apptainer/checkpoint/criu/<name>/img`用于存储CRIU检查点。如果指定了`--mem-dir`（memory directory的缩写）参数，则会在内存文件系统tmpfs中创建一个目录，以提高检查点读写的速度。

由于用户接口程序singularity没有特权，无法直接挂载tmpfs。因此，在本文中，选择在许多Linux发行版中默认提供的tmpfs挂载点/dev/shm下创建目录。`/dev/shm`目录允许所有用户访问，因此可以创建`/dev/shm/$USER/.apptainer/<name>`目录。

- `apptainer checkpoint list` 列出所有检查点目录
- `apptainer checkpoint delete <name>` 删除检查点目录

### 容器启动相关命令

- `apptainer instance start [options] –criu-launch <checkpoint name> <image path> <instance name> [args...]` 启动需要检查点功能的容器实例。


options是容器启动时附带的参数，可以对启动的容器进行配置，比如-B指定绑定挂载的目录、--ipc创建ipc命名空间等等，instance name作为容器名，后续可以通过这个容器名定位到运行的容器实例；args是传递给镜像默认运行程序的参数。为了不影响原始的容器实例启动流程，额外新增了—criu-launch参数，表示启动一个后续允许进行检查点操作的容器实例。其参数值为checkpoint name，是预先通过checkpoint create命令创建的检查点名字。

启动容器时需要做一些额外的工作，包括：

1. 在Apptainer用户接口中，需要关闭从父进程（通常是shell进程）继承来的打开文件。因为Apptainer不会对打开的文件进行额外的处理，如果不关闭这些文件，最终它们会被容器内的工作进程所继承。通过mount命名空间和chroot调用，容器拥有和主机隔离的rootfs，在容器中无法访问这些主机上的文件路径。而在执行dump操作时，CRIU会导出打开文件表，并访问这些打开文件。然而，由于这些文件在容器内无法访问，这将导致dump操作失败。因此，为了确保操作成功，需要将这些打开文件关闭。
2. 在容器运行时starter-suid进行rootfs初始化时，需要绑定挂载CRIU相关的二进制文件，包括可执行文件和依赖的动态链接库，以及检查点目录。Linux的绑定挂载机制（Bind Mount）可以将目录或文件挂载到另一个指定路径下，所以可以通过将主机上的目录或文件挂载到rootfs上，从而在容器内也可以访问到它们。绑定挂载CRIU相关文件是因为有些镜像中可能没有封装CRIU，必须保证任意镜像启动后在rootfs内都带有CRIU程序。检查点目录是上一节提到的、使用singularity checkpoint create命令创建的容器检查点目录，存储在共享的分布式文件系统中。这个目录将被绑定挂载到容器内的/.checkpoint目录下，同时，CRIU生成的镜像会存储在/.checkpoint/img目录中。如果在创建目录时设置了—mem-dir选项，则内存文件系统中的目录也会被绑定挂载到容器的/.checkpoint/img目录下。在这种情况下，CRIU将在主机内存中读写检查点文件，从而获得加速。
3. 给容器内进程创建终端会话。在原始流程中，Apptainer容器运行时starter-suid会让容器内的init进程的父进程master（它在容器启动后作为监控进程存在）创建终端会话。这样，容器内所有后续产生的进程都在master进程的终端会话内，master进程作为它们的session leader。但是，在执行dump操作时，CRIU需要获取session leader进程的信息。由于容器通过pid命名空间对进程进行了隔离，因此无法访问到master进程，并会报错：“A session leader of xxx is outside of its pid namespace”。因此，需要在给容器内的init进程创建一个终端会话，以便容器内进程的session leader也在容器的pid命名空间中。
4. 导出容器配置。容器的配置由两部分，一是启动容器时指定的参数选项，二是全局配置。这里只需导出启动容器时指定的参数即可，导出到容器检查点目录下，可供后续容器恢复时读取导入。
5. 设置容器网络。在容器技术中，为了让容器能够具有独立的IP地址并且与外部通信，需要设置容器网络。为此，需要在容器内部创建网络命名空间，然后在其中创建macvlan设备，并为该设备配置IP地址、子网掩码、默认网关等网络信息。为了实现这个目标，可以使用macvlan CNI插件。CNI（Container Network Interface）插件提供了一个标准的网络配置接口，使得各种容器运行时可以使用不同的网络方案进行网络配置。为了使用macvlan CNI插件配置容器网络，需要在容器检查点目录下放置相应的网络配置文件。用户可以更改这些配置文件中的IP地址、子网掩码和默认网关等网络信息。这些配置信息将被传递给macvlan CNI插件，并与容器所在的网络命名空间一同使用。通过这种方式，可以为容器提供独立的IP地址，并保证容器与外部网络在第二层被打通。容器重启时，会读取配置来重新设置容器网络，恢复成之前的状态。
6. 控制容器内工作进程的pid号为一个较大值。可以通过设置`/proc/sys/kernel/ns_last_pid`文件实现。ns_last_pid是Linux 3.3开始支持的特殊文件，包含内核分配的最后一个pid。当内核需要分配一个新的pid时，默认分配last_pid+1。因此，可以通过写入这个文件来配置当前pid命名空间内下一个生成的进程号。容器内的pid从1开始分配，因为实例模式默认开启pid命名空间。容器内有一些固定流程来进行初始化工作，会消耗固定数量的pid。因此，当实际启动工作进程时，它的pid会是一个固定的值。当从检查点恢复的容器启动时，CRIU进程作为工作进程，执行restore操作。它会和待恢复的工作进程为同一个pid。但是，CRIU恢复进程需要恢复其pid号，而这个pid号已经被CRIU自己占用了，会导致恢复失败。为了避免这个问题，需要控制容器内的工作进程的pid为一个较大值，默认为1024，即给`/proc/sys/kernel/ns_last_pid`文件写入1024。这样后续恢复容器时，启动的CRIU进程号不会和它冲突。
7. 记录应用的进程号。将工作进程的pid写到容器检查点目录的文件里，在后续进行CRIU dump操作时，需要将pid作为参数传递给CRIU。

### 检查点生成和恢复相关命令

- `apptainer checkpoint instance –criu [--page-server] <name>` 执行检查点生成操作。

--criu参数表明最终检查点生成的过程将通过CRIU来完成，与DMTCP作区分。--page-server选项用于控制CRIU dump调用是否带有--page-server参数。这个参数可以带一个目的端的地址，通过--address参数指定。—page-server选项在无盘迁移流程中起到了关键作用，因为它提醒CRIU在另一个节点上启动着CRIU page-server，导出检查点文件时，需要将内存相关数据直接传递给该page-server，而不是保存到本地的文件系统中。

- `apptainer checkpoint instance --criu –restore <name>` 执行检查点恢复操作

最终会在容器内调用 criu restore 来执行恢复动作。这个命令主要在无盘迁移流程中使用。

总体来说，容器内调用CRIU有三个步骤：

1. 在容器运行时，为了保证成功调用CRIU，需要在调用前给starter-suid设置适当的权限。使用capability进行授权，并将其切换成调用命令的普通用户。具体而言，对于5.9版本及以上的内核，我们应该将权限设置为CAP_CHECKPOINT_RESTORE | CAP_NET_ADMIN | CAP_SETPCAP；而对于5.9版本以下的内核，应该将权限设置为CAP_SYS_ADMIN | CAP_NET_ADMIN CAP_SETPCAP。通过预先设置capability授权的方式，在不暴露过多权限的前提下，确保CRIU后续在容器内正常运行。
2. 调用CRIU。为了让CRIU能够在容器内运行，需要将进程移动到容器的命名空间中。在Linux操作系统中，可以通过setns系统调用来实现命名空间的切换。切换之后进程在逻辑上就属于容器内。
3. 会根据Singularity不同的命令选项来设置CRIU的参数并调用它。

### 容器重启

- `apptainer instance start [options] –criu-restart <checkpoint name> [--page-server] <image path> <instance name> [args...]`

重启容器和普通的启动容器大部分工作类似，也都会进行启动容器最核心的工作：进行配置的收集和验证，设置rootfs，创建命名空间，启动容器进程，并监控容器。

额外执行的步骤：

1. 在用户接口中，Apptainer会导入容器启动时导出的配置，这些配置存储在容器的检查点目录中。
2. 恢复容器网络。容器网络的恢复实际上是对其进行重建。与启动容器时类似，此过程涉及网络命名空间和macvlan虚拟设备的创建和配置，容器运行时会读取检查点目录中的配置文件，配置文件中包括了恢复容器所需的网络信息，如容器IP、子网掩码、默认网关等。
3. 与执行检查点操作时一样，需要在调用CRIU前给当前进程设置合适的capability，彻底舍弃root身份，变为普通用户。
4. 在容器内调用CRIU，CRIU作为容器内的第一个工作进程启动。并根据是否指定—page-server参数来调用不同的CRIU功能。

## 需要修改的文件

不同的环境可能有所差异，可能导致运行失败。可以尝试调整这些文件，以适应不同的环境。

- `etc/criu-conf.yaml`

这个文件里是需要bind mount进容器的CRIU相关二进制文件。`make install`会安装在`/usr/local/etc/apptainer/`。如果在容器内，CRIU找不到动态库，可以尝试修改这个文件，添加依赖。

- `internal/pkg/checkpoint/criu/commands.go`

这个文件里涉及到CRIU的命令行参数。可以调整这些参数来选择CRIU的用法。

- `~/.apptainer/checkpoint/criu/<instance name>/30_macvlan.conflist`

这个文件是启动macvlan网络的配置文件。`instance start`时可以通过`--macvlan`选项启用网络虚拟化功能，配置macvlan设备。用户可以修改其中的IP地址、子网掩码和默认网关等网络信息。
