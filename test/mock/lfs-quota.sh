#!/bin/bash

# 模拟 lfs quota 命令
# 用法: ./lfs-quota.sh quota -u <user> <filesystem>
#      ./lfs-quota.sh quota -g <group> <filesystem>

# 检查参数数量
if [ $# -lt 4 ]; then
    echo "Usage: $0 quota [-u <user>|-g <group>] <filesystem>"
    exit 1
fi

# 解析参数
COMMAND=$1
OPTION=$2
TARGET=$3
FILESYSTEM=$4

# 检查命令格式
if [ "$COMMAND" != "quota" ]; then
    echo "Error: Invalid command format"
    echo "Usage: $0 quota [-u <user>|-g <group>] <filesystem>"
    exit 1
fi

# 根据选项类型处理
case "$OPTION" in
    "-u")
        # 用户配额查询
        case "$TARGET" in
            "wuwy")
                case "$FILESYSTEM" in
                    "/fs2")
                        echo "Disk quotas for usr wuwy (uid 30645):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "           /fs2 25678912  268435456 536870912       -  156789   500000 1000000       -"
                        ;;
                    "/fs1")
                        echo "Disk quotas for usr wuwy (uid 30645):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "           /fs1 12456780  134217728 268435456       -  78901   250000  500000       -"
                        ;;
                    *)
                        echo "Disk quotas for usr wuwy (uid 30645):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "         $FILESYSTEM        0         0        0       -       0        0       0       -"
                        ;;
                esac
                ;;
            "admin")
                case "$FILESYSTEM" in
                    "/fs2")
                        echo "Disk quotas for usr admin (uid 1001):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "           /fs2 51200000 1073741824 2147483648       -  300000  2000000 4000000       -"
                        ;;
                    *)
                        echo "Disk quotas for usr admin (uid 1001):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "         $FILESYSTEM 25600000  536870912 1073741824       -  150000  1000000 2000000       -"
                        ;;
                esac
                ;;
            *)
                echo "Disk quotas for usr $TARGET (uid 8888):"
                echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                echo "         $FILESYSTEM        0   52428800 104857600       -       0    50000  100000       -"
                ;;
        esac
        ;;
    "-g")
        # 组配额查询
        case "$TARGET" in
            "wuwy")
                case "$FILESYSTEM" in
                    "/fs2")
                        echo "Disk quotas for grp wuwy (gid 30645):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "           /fs2 37545936  536870912 1073741824       -  210311  1000000 2000000       -"
                        ;;
                    "/fs1")
                        echo "Disk quotas for grp wuwy (gid 30645):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "           /fs1 15672840  268435456 536870912       -  89245   500000 1000000       -"
                        ;;
                    *)
                        echo "Disk quotas for grp wuwy (gid 30645):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "         $FILESYSTEM        0         0        0       -       0        0       0       -"
                        ;;
                esac
                ;;
            "admin")
                case "$FILESYSTEM" in
                    "/fs2")
                        echo "Disk quotas for grp admin (gid 1001):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "           /fs2 102400000 2147483648 4294967296       -  500000  5000000 10000000       -"
                        ;;
                    *)
                        echo "Disk quotas for grp admin (gid 1001):"
                        echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                        echo "         $FILESYSTEM 51200000 1073741824 2147483648       -  250000  2500000 5000000       -"
                        ;;
                esac
                ;;
            *)
                echo "Disk quotas for grp $TARGET (gid 9999):"
                echo "     Filesystem  kbytes   quota   limit   grace   files   quota   limit   grace"
                echo "         $FILESYSTEM        0  104857600 209715200       -       0   100000  200000       -"
                ;;
        esac
        ;;
    *)
        echo "Error: Invalid option '$OPTION'"
        echo "Usage: $0 quota [-u <user>|-g <group>] <filesystem>"
        exit 1
        ;;
esac