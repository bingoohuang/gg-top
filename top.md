# TOP  信息解读[^1]

```sh
top - 17:45:14 up 37 min,  1 user,  load average: 0.00, 0.00, 0.00
当前时间：17：45分，运行37分钟，1个登录用户，1,5,15分钟的系统平均负载
Tasks:  79 total,   1 running,  41 sleeping,   0 stopped,   0 zombie
进程：总进程数，1个运行，41个睡眠中，0中止，0僵尸进程
%Cpu(s):  0.2 us,  0.0 sy,  0.0 ni, 99.0 id,  0.0 wa,  0.0 hi,  0.0 si,  0.8 st
CPU资源：
  us: is meaning of "user CPU time" 用户空间占用CPU百分比
  sy: is meaning of "system CPU time" 内核空间占用CPU百分比
  ni: is meaning of "nice CPU time" 用户进程空间内改变过优先级的进程占用CPU百分比
  id: is meaning of "idle" 空闲CPU百分比
  wa: is meaning of "iowait"  等待输入输出的CPU时间百分比
  hi：is meaning of "hardware irq" 硬件中断
  si : is meaning of "software irq" 软件中断
  st : is meaning of "steal time" ST为0表示流畅，CPU资源足够完全不需要等待，当数值增加时，表示服务器资源不够，母机可能超售。你的虚拟VPS需要等待分配物理CPU的等待时间的百分比，你排队等候分配资源的百分比。
KiB Mem :  2679236 total,  2465080 free,    99964 used,   114192 buff/cache
物理内存：总内存，空闲内存，使用中的内存，用作内核缓存的内存。
KiB Swap:   262140 total,   262140 free,        0 used.  2439240 avail Mem
虚拟内存交换区：总交换区，空闲交换区，使用中的，缓冲的交换区总量。
PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
PID — 进程id
USER — 进程所有者
PR — 进程优先级
NI — nice值。负值表示高优先级，正值表示低优先级
VIRT — 进程使用的虚拟内存总量，单位kb。VIRT=SWAP+RES
RES — 进程使用的、未被换出的物理内存大小，单位kb。RES=CODE+DATA
SHR — 共享内存大小，单位kb
S — 进程状态。D=不可中断的睡眠状态 R=运行 S=睡眠 T=跟踪/停止 Z=僵尸进程
%CPU — 上次更新到现在的CPU时间占用百分比
%MEM — 进程使用的物理内存百分比
TIME+ — 进程占用的CPU时间总计，单位1/100秒
COMMAND — 进程名称（命令名/命令行）
```

列名|含义
---|---
PID|进程id
PPID|父进程id
RUSER|Real user name
UID|进程所有者的用户id
USER|进程所有者的用户名
GROUP|进程所有者的组名
TTY|启动进程的终端名。不是从终端启动的进程则显示为 ?
PR|优先级
NI|nice值。负值表示高优先级，正值表示低优先级
P|最后使用的CPU，仅在多CPU环境下有意义
%CPU|上次更新到现在的CPU时间占用百分比
TIME|进程使用的CPU时间总计，单位秒
TIME+|进程使用的CPU时间总计，单位1/100秒
%MEM|进程使用的物理内存百分比
VIRT|进程使用的虚拟内存总量，单位kb。VIRT=SWAP+RES
SWAP|进程使用的虚拟内存中，被换出的大小，单位kb。
RES|进程使用的、未被换出的物理内存大小，单位kb。RES=CODE+DATA
CODE|可执行代码占用的物理内存大小，单位kb
DATA|可执行代码以外的部分(数据段+栈)占用的物理内存大小，单位kb
SHR|共享内存大小，单位kb
nFLT|页面错误次数
nDRT|最后一次写入到现在，被修改过的页面数。
S|进程状态： D=不可中断的睡眠状态，R=运行，S=睡眠，T=跟踪/停止，Z=僵尸进程
COMMAND|命令名/命令行
WCHAN|若该进程在睡眠，则显示睡眠中的系统函数名
Flags|任务标志，参考 sched.h

## What do VIRT, RES and SHR mean in the top command? [^2] 

- VIRT: 虚拟空间大小
- RES:  占用的物理内存
- SHR:  和其他进程共享的内存 [^3]

It can be seen from man top[^4] in terminal as

[^2]: https://askubuntu.com/a/176002
[^4]: http://manpages.ubuntu.com/manpages/precise/en/man1/top.1.html
[^3]: https://www.orchome.com/298
[^1]: https://www.pigji.com/776.html

DESCRIPTIONS of Fields Listed below are top's available fields. They are always associated with the letter shown, regardless of the position you may have established for them with the 'o' (Order fields) interactive command.

```sh
Any field is selectable as the sort field, and you control whether
they  are  sorted  high-to-low  or  low-to-high.   For  additional
information on sort provisions see topic 3c. TASK Area Commands.

a: PID  --  Process Id
The  task's unique process ID, which periodically wraps, though
never restarting at zero.

b: PPID  --  Parent Process Pid
The process ID of a task's parent.

c: RUSER  --  Real User Name
The real user name of the task's owner.

d: UID  --  User Id
The effective user ID of the task's owner.

e: USER  --  User Name
The effective user name of the task's owner.

f: GROUP  --  Group Name
The effective group name of the task's owner.

g: TTY  --  Controlling Tty
The name of the controlling  terminal.   This  is  usually  the
device  (serial  port,  pty,  etc.)  from which the process was
started, and which it uses for input  or  output.   However,  a
task  need  not  be  associated  with a terminal, in which case
you'll see '?' displayed.

h: PR  --  Priority
The priority of the task.

i: NI  --  Nice value
The nice value of the task.  A negative nice value means higher
priority,  whereas  a positive nice value means lower priority.
Zero in this field simply means priority will not  be  adjusted
in determining a task's dispatchability.

j: P  --  Last used CPU (SMP)
A  number  representing the last used processor.  In a true SMP
environment this will likely change frequently since the kernel
intentionally  uses  weak  affinity.   Also,  the  very  act of
running top  may  break  this  weak  affinity  and  cause  more
processes  to  change  CPUs  more  often  (because of the extra
demand for cpu time).

k: %CPU  --  CPU usage
The task's share of the elapsed CPU time since the last  screen
update, expressed as a percentage of total CPU time.  In a true
SMP environment, if 'Irix mode' is Off,  top  will  operate  in
'Solaris  mode' where a task's cpu usage will be divided by the
total number of CPUs.  You toggle 'Irix/Solaris' modes with the
'I' interactive command.

l: TIME  --  CPU Time
Total  CPU  time  the  task  has  used  since it started.  When
'Cumulative mode' is On, each process is listed  with  the  cpu
time  that  it  and  its  dead  children  has used.  You toggle
'Cumulative mode' with 'S', which is a command-line option  and
an  interactive  command.   See the 'S' interactive command for
additional information regarding this mode.

m: TIME+  --  CPU Time, hundredths
The same as 'TIME', but  reflecting  more  granularity  through
hundredths of a second.

n: %MEM  --  Memory usage (RES)
A task's currently used share of available physical memory.

o: VIRT  --  Virtual Image (kb)
The  total  amount  of  virtual  memory  used  by the task.  It
includes all code, data and shared libraries  plus  pages  that
have  been  swapped out and pages that have been mapped but not
used.

p: SWAP  --  Swapped size (kb)
Memory that is not resident but is present in a task.  This  is
memory  that  has been swapped out but could include additional
non-resident memory.  This column is calculated by  subtracting
physical memory from virtual memory.

q: RES  --  Resident size (kb)
The non-swapped physical memory a task has used.

r: CODE  --  Code size (kb)
The  amount  of virtual memory devoted to executable code, also
known as the 'text resident set' size or TRS.

s: DATA  --  Data+Stack size (kb)
The amount of virtual memory devoted to other  than  executable
code, also known as the 'data resident set' size or DRS.

t: SHR  --  Shared Mem size (kb)
The amount of shared memory used by a task.  It simply reflects
memory that could be potentially shared with other processes.

u: nFLT  --  Page Fault count
The number of major page faults that have occurred for a  task.
A  page  fault  occurs  when a process attempts to read from or
write to a virtual page that is not currently  present  in  its
address  space.   A  major  page  fault is when backing storage
access (such as  a  disk)  is  involved  in  making  that  page
available.

v: nDRT  --  Dirty Pages count
The  number  of  pages  that have been modified since they were
last written to disk.  Dirty pages  must  be  written  to  disk
before  the  corresponding physical memory location can be used
for some other virtual page.

w: S  --  Process Status
The status of the task which can be one of:
'D' = uninterruptible sleep
'R' = running
'S' = sleeping
'T' = traced or stopped
'Z' = zombie

      Tasks shown as running should be more properly  thought  of  as
      'ready  to run'  --  their task_struct is simply represented on
      the Linux run-queue.  Even without a true SMP machine, you  may
      see  numerous  tasks  in  this  state  depending on top's delay
      interval and nice value.
```

For CPU

2c. CPU States The CPU states are shown in the Summary Area. They are always shown as a percentage and are for the time between now and the last refresh.

```sh
    us  --  User CPU time
      The time the CPU has spent running users'  processes  that  are
      not niced.

    sy  --  System CPU time
      The  time  the  CPU  has  spent  running  the  kernel  and  its
      processes.

    ni  --  Nice CPU time
      The time the CPU has spent running users'  proccess  that  have
      been niced.

    wa  --  iowait
      Amount of time the CPU has been waiting for I/O to complete.

    hi  --  Hardware IRQ
      The  amount  of  time  the  CPU  has  been  servicing  hardware
      interrupts.

    si  --  Software Interrupts
      The  amount  of  time  the  CPU  has  been  servicing  software
      interrupts.

    st  --  Steal Time
      The  amount  of  CPU  'stolen' from this virtual machine by the
      hypervisor for other tasks (such  as  running  another  virtual
      machine)
```

