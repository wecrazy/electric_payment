#!/bin/bash

# Script to switch between dev and prod config modes

CONFIG_FILE="config/conf.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Config file '$CONFIG_FILE' not found!"
    exit 1
fi

echo "Current CONFIG_MODE:"
grep "^CONFIG_MODE:" "$CONFIG_FILE"

# Check if argument is provided (for non-interactive use)
if [ -n "$1" ]; then
    case $1 in
        dev|1)
            sed -i 's/CONFIG_MODE: "prod"/CONFIG_MODE: "dev"/g' "$CONFIG_FILE"
            echo "✅ Switched to development mode"
            ;;
        prod|2)
            sed -i 's/CONFIG_MODE: "dev"/CONFIG_MODE: "prod"/g' "$CONFIG_FILE"
            echo "✅ Switched to PRODUCTION mode"
            ;;
        *)
            echo "❌ Invalid argument. Use 'dev' or 'prod' (or 1 or 2)"
            exit 1
            ;;
    esac
else
    # Interactive mode
    echo ""
    echo "Select mode:"
    echo "1) Development (dev)"
    echo "2) Production (prod)"
    read -p "Enter choice (1 or 2): " choice

    case $choice in
        1)
            sed -i 's/CONFIG_MODE: "prod"/CONFIG_MODE: "dev"/g' "$CONFIG_FILE"
            echo "✅ Switched to development mode"
            ;;
        2)
            sed -i 's/CONFIG_MODE: "dev"/CONFIG_MODE: "prod"/g' "$CONFIG_FILE"
            echo "✅ Switched to PRODUCTION mode"
            ;;
        *)
            echo "❌ Invalid choice. Exiting."
            exit 1
            ;;
    esac
fi

echo ""
echo "New CONFIG_MODE:"
grep "^CONFIG_MODE:" "$CONFIG_FILE"