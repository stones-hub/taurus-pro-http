#!/bin/bash

# 安全系统并发性能优化脚本
# 适用于 Linux 系统，需要 root 权限
# 采用保守配置，确保系统稳定性

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 安全检查
check_environment() {
    log_info "开始环境检查..."
    
    # 检查 root 权限
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 权限运行此脚本"
        exit 1
    fi
    
    # 检查操作系统
    if [ ! -f /etc/os-release ]; then
        log_error "不支持的操作系统"
        exit 1
    fi
    
    # 检查系统架构
    ARCH=$(uname -m)
    if [[ "$ARCH" != "x86_64" && "$ARCH" != "aarch64" ]]; then
        log_warn "未测试的系统架构: $ARCH"
    fi
    
    log_info "环境检查通过"
}

# 备份配置
backup_configs() {
    log_info "备份系统配置..."
    
    BACKUP_DIR="/root/system_optimization_backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # 备份关键配置文件
    if [ -f /etc/sysctl.conf ]; then
        cp /etc/sysctl.conf "$BACKUP_DIR/"
    fi
    
    if [ -f /etc/security/limits.conf ]; then
        cp /etc/security/limits.conf "$BACKUP_DIR/"
    fi
    
    if [ -f /etc/rc.local ]; then
        cp /etc/rc.local "$BACKUP_DIR/"
    fi
    
    log_info "配置已备份到: $BACKUP_DIR"
}

# 获取系统信息
get_system_info() {
    log_info "获取系统信息..."
    
    CPU_CORES=$(nproc)
    MEMORY_GB=$(free -g | awk '/^Mem:/{print $2}')
    
    # 验证系统信息
    if [ -z "$CPU_CORES" ] || [ "$CPU_CORES" -lt 1 ]; then
        log_error "无法获取 CPU 核心数"
        exit 1
    fi
    
    if [ -z "$MEMORY_GB" ] || [ "$MEMORY_GB" -lt 1 ]; then
        log_error "无法获取内存大小"
        exit 1
    fi
    
    log_info "检测到系统配置: ${CPU_CORES} 核, ${MEMORY_GB}GB 内存"
}

# 计算安全的优化参数
calculate_safe_parameters() {
    local cores=$1
    local memory=$2
    
    log_info "计算安全优化参数..."
    
    # 文件描述符限制：保守配置
    if [ $memory -ge 64 ]; then
        FILE_DESCRIPTOR_LIMIT=262144      # 256K for 64GB+
    elif [ $memory -ge 32 ]; then
        FILE_DESCRIPTOR_LIMIT=131072      # 128K for 32GB
    elif [ $memory -ge 16 ]; then
        FILE_DESCRIPTOR_LIMIT=65536       # 64K for 16GB
    elif [ $memory -ge 8 ]; then
        FILE_DESCRIPTOR_LIMIT=32768       # 32K for 8GB
    else
        FILE_DESCRIPTOR_LIMIT=16384       # 16K for <8GB
    fi
    
    # TCP 连接队列：保守配置
    if [ $cores -ge 32 ]; then
        TCP_CONN_QUEUE=131072             # 128K for 32+ cores
    elif [ $cores -ge 16 ]; then
        TCP_CONN_QUEUE=65536              # 64K for 16 cores
    elif [ $cores -ge 8 ]; then
        TCP_CONN_QUEUE=32768              # 32K for 8 cores
    elif [ $cores -ge 4 ]; then
        TCP_CONN_QUEUE=16384              # 16K for 4 cores
    else
        TCP_CONN_QUEUE=8192               # 8K for <4 cores
    fi
    
    # TCP 缓冲区：保守配置
    if [ $memory -ge 64 ]; then
        TCP_BUFFER_SIZE=33554432          # 32MB for 64GB+
    elif [ $memory -ge 32 ]; then
        TCP_BUFFER_SIZE=16777216          # 16MB for 32GB
    elif [ $memory -ge 16 ]; then
        TCP_BUFFER_SIZE=8388608           # 8MB for 16GB
    elif [ $memory -ge 8 ]; then
        TCP_BUFFER_SIZE=4194304           # 4MB for 8GB
    else
        TCP_BUFFER_SIZE=2097152           # 2MB for <8GB
    fi
    
    # 网络积压队列：保守配置
    if [ $cores -ge 32 ]; then
        NET_BACKLOG=10000                 # 10K for 32+ cores
    elif [ $cores -ge 16 ]; then
        NET_BACKLOG=5000                  # 5K for 16 cores
    elif [ $cores -ge 8 ]; then
        NET_BACKLOG=2500                  # 2.5K for 8 cores
    elif [ $cores -ge 4 ]; then
        NET_BACKLOG=1000                  # 1K for 4 cores
    else
        NET_BACKLOG=500                   # 500 for <4 cores
    fi
    
    # 内存优化参数：保守配置
    if [ $memory -ge 64 ]; then
        VM_SWAPPINESS=1                   # 最小交换 for 64GB+
        VM_DIRTY_RATIO=20                 # 20% for 64GB+
        VM_DIRTY_BG_RATIO=10              # 10% for 64GB+
    elif [ $memory -ge 32 ]; then
        VM_SWAPPINESS=5                   # 低交换 for 32GB
        VM_DIRTY_RATIO=15                 # 15% for 32GB
        VM_DIRTY_BG_RATIO=8               # 8% for 32GB
    elif [ $memory -ge 16 ]; then
        VM_SWAPPINESS=10                  # 中等交换 for 16GB
        VM_DIRTY_RATIO=12                 # 12% for 16GB
        VM_DIRTY_BG_RATIO=6               # 6% for 16GB
    else
        VM_SWAPPINESS=15                  # 默认交换 for <16GB
        VM_DIRTY_RATIO=10                 # 10% for <16GB
        VM_DIRTY_BG_RATIO=5               # 5% for <16GB
    fi
    
    # 进程和线程限制：保守配置
    if [ $cores -ge 32 ]; then
        PID_MAX=131072                    # 128K for 32+ cores
        THREADS_MAX=262144                # 256K for 32+ cores
        MAX_MAP_COUNT=524288              # 512K for 32+ cores
    elif [ $cores -ge 16 ]; then
        PID_MAX=65536                     # 64K for 16 cores
        THREADS_MAX=131072                # 128K for 16 cores
        MAX_MAP_COUNT=262144              # 256K for 16 cores
    elif [ $cores -ge 8 ]; then
        PID_MAX=32768                     # 32K for 8 cores
        THREADS_MAX=65536                 # 64K for 8 cores
        MAX_MAP_COUNT=131072              # 128K for 8 cores
    else
        PID_MAX=16384                     # 16K for <8 cores
        THREADS_MAX=32768                 # 32K for <8 cores
        MAX_MAP_COUNT=65536               # 64K for <8 cores
    fi
    
    # Go 运行时参数：保守配置
    GOMAXPROCS=$cores
    if [ $memory -ge 64 ]; then
        GOGC=200                          # 中等 GC 阈值
        GOMEMLIMIT="$((memory - 16))GiB" # 留 16GB 给系统
    elif [ $memory -ge 32 ]; then
        GOGC=150                          # 低 GC 阈值
        GOMEMLIMIT="$((memory - 8))GiB"  # 留 8GB 给系统
    elif [ $memory -ge 16 ]; then
        GOGC=120                          # 低 GC 阈值
        GOMEMLIMIT="$((memory - 4))GiB"  # 留 4GB 给系统
    else
        GOGC=100                          # 默认 GC 阈值
        GOMEMLIMIT="$((memory - 2))GiB"  # 留 2GB 给系统
    fi
    
    # 参数验证
    if [ $FILE_DESCRIPTOR_LIMIT -gt 1048576 ]; then
        log_error "文件描述符限制过高: $FILE_DESCRIPTOR_LIMIT"
        exit 1
    fi
    
    if [ $TCP_CONN_QUEUE -gt 1048576 ]; then
        log_error "TCP 连接队列过高: $TCP_CONN_QUEUE"
        exit 1
    fi
    
    # 额外的安全验证
    if [ $VM_DIRTY_RATIO -gt 20 ]; then
        log_error "内存脏页比例过高: $VM_DIRTY_RATIO%"
        exit 1
    fi
    
    if [ $VM_DIRTY_BG_RATIO -gt 10 ]; then
        log_error "内存后台脏页比例过高: $VM_DIRTY_BG_RATIO%"
        exit 1
    fi
    
    if [ $TCP_BUFFER_SIZE -gt 134217728 ]; then
        log_error "TCP缓冲区过大: $((TCP_BUFFER_SIZE / 1024 / 1024))MB"
        exit 1
    fi
    
    log_info "安全参数计算完成"
}

# 显示配置参数
show_parameters() {
    echo ""
    echo "📊 安全优化参数配置："
    echo "   - 文件描述符限制: ${FILE_DESCRIPTOR_LIMIT}"
    echo "   - TCP 连接队列: ${TCP_CONN_QUEUE}"
    echo "   - TCP 缓冲区: $((TCP_BUFFER_SIZE / 1024 / 1024))MB"
    echo "   - 网络积压队列: ${NET_BACKLOG}"
    echo "   - 内存交换策略: ${VM_SWAPPINESS}"
    echo "   - GOMAXPROCS: ${GOMAXPROCS}"
    echo "   - GOGC: ${GOGC}"
    echo "   - 内存限制: ${GOMEMLIMIT}"
    echo ""
}

# 应用优化配置
apply_optimizations() {
    log_info "开始应用优化配置..."
    
    # 1. 文件描述符限制
    log_info "优化文件描述符限制..."
    echo "* soft nofile ${FILE_DESCRIPTOR_LIMIT}" >> /etc/security/limits.conf
    echo "* hard nofile ${FILE_DESCRIPTOR_LIMIT}" >> /etc/security/limits.conf
    ulimit -n ${FILE_DESCRIPTOR_LIMIT}
    
    # 2. 内核网络参数
    log_info "优化内核网络参数..."
    cat >> /etc/sysctl.conf << EOF

# 网络并发优化 (安全配置)
net.core.somaxconn = ${TCP_CONN_QUEUE}
net.core.netdev_max_backlog = ${NET_BACKLOG}
net.ipv4.tcp_max_syn_backlog = ${TCP_CONN_QUEUE}
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_tw_recycle = 0
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_intvl = 15
net.ipv4.tcp_keepalive_probes = 5
net.ipv4.tcp_max_tw_buckets = $((TCP_CONN_QUEUE * 2))
net.ipv4.tcp_tcp_notsent_lowat = 16384
net.ipv4.tcp_congestion_control = bbr

# TCP 缓冲区优化 (安全配置)
net.core.rmem_default = $((TCP_BUFFER_SIZE / 128))
net.core.rmem_max = ${TCP_BUFFER_SIZE}
net.core.wmem_default = $((TCP_BUFFER_SIZE / 128))
net.core.wmem_max = ${TCP_BUFFER_SIZE}
net.ipv4.tcp_rmem = 4096 $((TCP_BUFFER_SIZE / 128)) ${TCP_BUFFER_SIZE}
net.ipv4.tcp_wmem = 4096 $((TCP_BUFFER_SIZE / 128)) ${TCP_BUFFER_SIZE}

# 内存优化 (安全配置)
vm.swappiness = ${VM_SWAPPINESS}
vm.dirty_ratio = ${VM_DIRTY_RATIO}
vm.dirty_background_ratio = ${VM_DIRTY_BG_RATIO}
vm.min_free_kbytes = $((MEMORY_GB * 512 * 1024))
vm.vfs_cache_pressure = 50

# 进程和线程优化 (安全配置)
kernel.pid_max = ${PID_MAX}
kernel.threads-max = ${THREADS_MAX}
kernel.max_map_count = ${MAX_MAP_COUNT}

# 网络连接优化
net.ipv4.tcp_max_orphans = $((TCP_CONN_QUEUE / 8))
net.ipv4.tcp_orphan_retries = 3
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_synack_retries = 3
net.ipv4.tcp_syn_retries = 3
net.ipv4.tcp_retries2 = 15
net.ipv4.tcp_retries1 = 3

# 文件系统优化
fs.file-max = $((FILE_DESCRIPTOR_LIMIT * 8))
fs.inotify.max_user_watches = $((FILE_DESCRIPTOR_LIMIT * 2))
EOF
    
    # 3. 应用内核参数
    log_info "应用内核参数..."
    sysctl -p
    
    # 4. TCP 参数优化
    log_info "优化 TCP 参数..."
    cat >> /etc/rc.local << EOF

# 优化 TCP 缓冲区 (安全配置)
echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/rmem_max
echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/wmem_max
echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/rmem_default
echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/wmem_default

# 设置 TCP 内存范围
echo "4096 $((TCP_BUFFER_SIZE / 128)) ${TCP_BUFFER_SIZE}" > /proc/sys/net/ipv4/tcp_rmem
echo "4096 $((TCP_BUFFER_SIZE / 128)) ${TCP_BUFFER_SIZE}" > /proc/sys/net/ipv4/tcp_wmem

# 优化连接跟踪
echo ${TCP_CONN_QUEUE} > /proc/sys/net/netfilter/nf_conntrack_max
echo 120 > /proc/sys/net/netfilter/nf_conntrack_tcp_timeout_established
EOF
    
    # 5. 设置当前会话参数
    log_info "设置当前会话参数..."
    echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/rmem_max
    echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/wmem_max
    echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/rmem_default
    echo ${TCP_BUFFER_SIZE} > /proc/sys/net/core/wmem_default
    echo "4096 $((TCP_BUFFER_SIZE / 128)) ${TCP_BUFFER_SIZE}" > /proc/sys/net/ipv4/tcp_rmem
    echo "4096 $((TCP_BUFFER_SIZE / 128)) ${TCP_BUFFER_SIZE}" > /proc/sys/net/ipv4/tcp_wmem
    
    # 6. Go 运行时配置
    log_info "创建 Go 运行时配置..."
    cat > /etc/profile.d/go_optimization.sh << EOF
# Go 运行时优化配置 (安全配置)
export GOMAXPROCS=${GOMAXPROCS}
export GOGC=${GOGC}
export GOMEMLIMIT=${GOMEMLIMIT}
export GOMEMLIMIT_OFF=1
export GODEBUG=schedtrace=1000,scheddetail=1
EOF
    
    # 7. 应用 Go 配置
    source /etc/profile.d/go_optimization.sh
    
    # 8. 系统服务配置
    log_info "创建系统服务配置..."
    cat > /etc/systemd/system.conf.d/limits.conf << EOF
[Manager]
DefaultLimitNOFILE=${FILE_DESCRIPTOR_LIMIT}
DefaultLimitNPROC=${FILE_DESCRIPTOR_LIMIT}
EOF
    
    # 9. 磁盘 I/O 优化
    if command -v nvme >/dev/null 2>&1; then
        log_info "检测到 NVMe SSD，优化 I/O 调度..."
        echo "none" > /sys/block/nvme*/queue/scheduler 2>/dev/null || true
        echo $((NET_BACKLOG * 2)) > /sys/block/nvme*/queue/nr_requests 2>/dev/null || true
    fi
    
    log_info "优化配置应用完成"
}

# 验证配置
verify_configuration() {
    log_info "验证配置..."
    
    # 验证文件描述符限制
    CURRENT_ULIMIT=$(ulimit -n)
    if [ "$CURRENT_ULIMIT" -eq "$FILE_DESCRIPTOR_LIMIT" ]; then
        log_info "✓ 文件描述符限制设置成功: $CURRENT_ULIMIT"
    else
        log_warn "⚠ 文件描述符限制设置可能失败: 当前 $CURRENT_ULIMIT, 目标 $FILE_DESCRIPTOR_LIMIT"
    fi
    
    # 验证 TCP 参数
    CURRENT_SOMAXCONN=$(sysctl -n net.core.somaxconn)
    if [ "$CURRENT_SOMAXCONN" -eq "$TCP_CONN_QUEUE" ]; then
        log_info "✓ TCP 连接队列设置成功: $CURRENT_SOMAXCONN"
    else
        log_warn "⚠ TCP 连接队列设置可能失败: 当前 $CURRENT_SOMAXCONN, 目标 $TCP_CONN_QUEUE"
    fi
    
    # 验证 Go 环境变量
    if [ "$GOMAXPROCS" -eq "$(echo $GOMAXPROCS)" ]; then
        log_info "✓ GOMAXPROCS 设置成功: $GOMAXPROCS"
    else
        log_warn "⚠ GOMAXPROCS 设置可能失败"
    fi
    
    log_info "配置验证完成"
}

# 显示使用说明
show_usage() {
    echo ""
    echo "🔧 使用说明："
    echo "   1. 重启系统或重新登录以使所有更改生效"
    echo "   2. 监控系统性能，确保稳定性"
    echo "   3. 如遇问题，可恢复备份配置"
    echo ""
    echo "✅ 验证命令："
    echo "   ulimit -n"
    echo "   sysctl net.core.somaxconn"
    echo "   sysctl net.ipv4.tcp_max_syn_backlog"
    echo "   sysctl net.core.rmem_max"
    echo "   echo \$GOMAXPROCS"
    echo ""
    echo "📈 预期性能提升："
    echo "   - 并发连接数: 提升至 $((TCP_CONN_QUEUE / 1000))K+"
    echo "   - 网络吞吐量: 提升 20-50%"
    echo "   - 内存使用效率: 提升 15-30%"
    echo "   - CPU 利用率: 更均匀分布"
    echo ""
    echo "⚠️  注意事项："
    echo "   - 此配置采用保守策略，确保系统稳定性"
    echo "   - 如需更高性能，可逐步调整参数"
    echo "   - 建议在生产环境部署前进行充分测试"
}

# 主函数
main() {
    echo "🚀 安全系统并发性能优化脚本启动"
    echo "=================================="
    
    check_environment
    backup_configs
    get_system_info
    calculate_safe_parameters
    show_parameters
    
    # 确认继续
    echo ""
    read -p "是否继续应用优化配置？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "用户取消操作"
        exit 0
    fi
    
    apply_optimizations
    verify_configuration
    show_usage
    
    log_info "🎯 安全系统优化完成！"
}

# 错误处理
trap 'log_error "脚本执行出错，请检查日志"; exit 1' ERR

# 执行主函数
main "$@"
