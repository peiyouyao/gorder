#!/bin/bash

# 项目根目录
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INTERNAL_DIR="$ROOT_DIR/internal"

# 模块列表
MODULES=("common" "order" "payment" "stock" "kitchen")

# 函数：在模块下执行 go mod tidy
tidy_module() {
    local module=$1
    local module_dir="$INTERNAL_DIR/$module"
    if [ -d "$module_dir" ]; then
        echo "🔄 Running 'go mod tidy' in $module_dir"
        (cd "$module_dir" && go mod tidy)
    else
        echo "❌ Module '$module' does not exist in $INTERNAL_DIR"
        exit 1
    fi
}

# 主逻辑
case "$1" in
    all)
        echo "📦 Tidying all modules..."
        for module in "${MODULES[@]}"; do
            tidy_module "$module"
        done
        ;;
    common|order|payment|stock|kitchen)
        tidy_module "$1"
        ;;
    *)
        echo "Usage: $0 [common|order|payment|stock|kitchen|all]"
        exit 1
        ;;
esac

echo "✅ Done."