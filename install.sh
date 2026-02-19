#!/bin/sh
set -e

REPO="KashifKhn/remove-comments"
BINARY_NAME="remove-comments"

CYAN='\033[0;36m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
MUTED='\033[0;2m'
BOLD='\033[1m'
NC='\033[0m'

requested_version=""
no_modify_path=false

usage() {
    cat <<EOF
remove-comments Installer - Remove all comments from source files

Usage: install.sh [options]

Options:
    -h, --help              Display this help message
    -v, --version <version> Install a specific version (e.g., 1.0.0)
        --no-modify-path    Don't modify shell config files

Examples:
    curl -fsSL https://raw.githubusercontent.com/$REPO/main/install.sh | sh
    curl -fsSL ... | sh -s -- --version 1.0.0
    curl -fsSL ... | sh -s -- --no-modify-path
EOF
}

print_logo() {
    printf "\n"
    printf "  ${CYAN}remove-comments${NC}\n"
    printf "\n"
    printf "  ${MUTED}Remove all comments from source files${NC}\n"
    printf "  ${MUTED}https://github.com/${REPO}${NC}\n"
    printf "\n"
}

print_success() {
    printf "  ${GREEN}✓${NC} $1\n"
}

print_error() {
    printf "  ${RED}✗${NC} $1\n"
}

print_info() {
    printf "  ${CYAN}→${NC} $1\n"
}

print_warning() {
    printf "  ${YELLOW}!${NC} $1\n"
}

get_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        *)              echo "unknown" ;;
    esac
}

detect_rosetta() {
    if [ "$(get_os)" = "darwin" ] && [ "$(get_arch)" = "amd64" ]; then
        rosetta_flag=$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)
        if [ "$rosetta_flag" = "1" ]; then
            echo "arm64"
            return
        fi
    fi
    get_arch
}

get_latest_version() {
    max_retries=3
    retry_delay=2
    attempt=1

    while [ $attempt -le $max_retries ]; do
        response=$(curl -sL -w "\n%{http_code}" "https://api.github.com/repos/${REPO}/releases/latest")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')

        if [ "$http_code" = "200" ]; then
            version=$(echo "$body" | grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/')
            if [ -n "$version" ]; then
                echo "$version"
                return 0
            fi
        fi

        if [ "$http_code" = "502" ] || [ "$http_code" = "503" ] || [ "$http_code" = "504" ]; then
            if [ $attempt -lt $max_retries ]; then
                sleep $retry_delay
            fi
            attempt=$((attempt + 1))
            continue
        fi

        break
    done

    echo ""
}

check_existing_version() {
    bin=""

    if command -v remove-comments >/dev/null 2>&1; then
        bin=$(command -v remove-comments)
    elif [ -x "/usr/local/bin/remove-comments" ]; then
        bin="/usr/local/bin/remove-comments"
    elif [ -x "$HOME/.local/bin/remove-comments" ]; then
        bin="$HOME/.local/bin/remove-comments"
    fi

    if [ -n "$bin" ] && [ -x "$bin" ]; then
        tmp=$(mktemp)
        "$bin" --version > "$tmp" 2>&1 || true
        if [ -s "$tmp" ]; then
            installed_version=$(sed -E 's/.*v?([0-9][0-9.]+).*/\1/' "$tmp")
            rm -f "$tmp"
            if [ -n "$installed_version" ]; then
                echo "$installed_version"
                return
            fi
        fi
        rm -f "$tmp"
    fi
    echo ""
}

print_progress() {
    bytes="$1"
    total="$2"

    if [ "$total" -le 0 ] 2>/dev/null; then
        return 0
    fi

    width=50
    percent=$((bytes * 100 / total))
    if [ "$percent" -gt 100 ]; then
        percent=100
    fi

    filled=$((percent * width / 100))
    empty=$((width - filled))

    bar=""
    i=0
    while [ $i -lt $filled ]; do
        bar="${bar}■"
        i=$((i + 1))
    done
    i=0
    while [ $i -lt $empty ]; do
        bar="${bar}･"
        i=$((i + 1))
    done

    printf "\r  ${CYAN}%s${NC} ${MUTED}%3d%%${NC}" "$bar" "$percent"
}

download_with_progress() {
    url="$1"
    output="$2"

    if [ ! -t 2 ]; then
        curl -fsSL "$url" -o "$output"
        return $?
    fi

    printf "\033[?25l"

    trap 'printf "\033[?25h"' EXIT INT TERM

    total_size=$(curl -sI "$url" | grep -i "content-length" | tail -1 | tr -d '\r' | awk '{print $2}')

    if [ -z "$total_size" ] || [ "$total_size" = "0" ]; then
        curl -fsSL "$url" -o "$output"
        ret=$?
        printf "\033[?25h"
        echo ""
        return $ret
    fi

    curl -fsSL "$url" -o "$output" 2>/dev/null &
    curl_pid=$!

    while kill -0 "$curl_pid" 2>/dev/null; do
        if [ -f "$output" ]; then
            current_size=$(wc -c < "$output" 2>/dev/null | tr -d ' ' || echo "0")
            print_progress "$current_size" "$total_size"
        fi
        sleep 0.1
    done

    wait $curl_pid
    ret=$?

    if [ $ret -eq 0 ]; then
        print_progress "$total_size" "$total_size"
    fi

    printf "\033[?25h"
    echo ""

    return $ret
}

add_to_path() {
    config_file="$1"
    path_command="$2"

    if [ ! -f "$config_file" ]; then
        return 1
    fi

    if grep -q "$INSTALL_DIR" "$config_file" 2>/dev/null; then
        return 0
    fi

    if [ -w "$config_file" ]; then
        echo "" >> "$config_file"
        echo "# remove-comments" >> "$config_file"
        echo "$path_command" >> "$config_file"
        print_success "Added to PATH in $config_file"
        return 0
    fi

    return 1
}

configure_path() {
    if [ "$no_modify_path" = true ]; then
        return
    fi

    case "$PATH" in
        *"$INSTALL_DIR"*)
            return
            ;;
    esac

    current_shell=$(basename "$SHELL" 2>/dev/null || echo "sh")

    path_added=false

    case "$current_shell" in
        fish)
            if add_to_path "$HOME/.config/fish/config.fish" "fish_add_path $INSTALL_DIR"; then
                path_added=true
            fi
            ;;
        zsh)
            for config_file in "$HOME/.zshrc" "$HOME/.zshenv"; do
                if [ -f "$config_file" ]; then
                    if add_to_path "$config_file" "export PATH=\"$INSTALL_DIR:\$PATH\""; then
                        path_added=true
                        break
                    fi
                fi
            done
            ;;
        bash)
            for config_file in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile"; do
                if [ -f "$config_file" ]; then
                    if add_to_path "$config_file" "export PATH=\"$INSTALL_DIR:\$PATH\""; then
                        path_added=true
                        break
                    fi
                fi
            done
            ;;
        *)
            if [ -f "$HOME/.profile" ]; then
                if add_to_path "$HOME/.profile" "export PATH=\"$INSTALL_DIR:\$PATH\""; then
                    path_added=true
                fi
            fi
            ;;
    esac

    if [ "$path_added" = false ]; then
        print_warning "Add to your PATH manually:"
        printf "    ${MUTED}export PATH=\"%s:\$PATH\"${NC}\n" "$INSTALL_DIR"
    fi
}

download_and_install() {
    VERSION="$1"
    OS="$2"
    ARCH="$3"

    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        print_error "Unsupported platform: OS=$(uname -s), Arch=$(uname -m)"
        exit 1
    fi

    if [ "$OS" = "windows" ]; then
        ARCHIVE="${BINARY_NAME}-${OS}-${ARCH}.zip"
    else
        ARCHIVE="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ARCHIVE}"

    print_info "Platform: ${BOLD}${OS}/${ARCH}${NC}"
    echo ""

    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    if ! download_with_progress "$DOWNLOAD_URL" "${TMP_DIR}/${ARCHIVE}"; then
        print_error "Failed to download ${BINARY_NAME} v${VERSION}"
        print_info "Check releases: https://github.com/${REPO}/releases"
        exit 1
    fi

    cd "$TMP_DIR"
    if [ "$OS" = "windows" ]; then
        unzip -q "$ARCHIVE" 2>/dev/null || {
            print_error "Failed to extract archive"
            exit 1
        }
    else
        tar -xzf "$ARCHIVE" 2>/dev/null || {
            print_error "Failed to extract archive"
            exit 1
        }
    fi

    if [ "$OS" = "windows" ]; then
        BINARY="${BINARY_NAME}-${OS}-${ARCH}.exe"
        TARGET_BINARY="${BINARY_NAME}.exe"
    else
        BINARY="${BINARY_NAME}-${OS}-${ARCH}"
        TARGET_BINARY="${BINARY_NAME}"
    fi

    INSTALL_DIR="/usr/local/bin"

    if [ ! -w "$INSTALL_DIR" ]; then
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi

    export INSTALL_DIR

    mv "$BINARY" "${INSTALL_DIR}/${TARGET_BINARY}"
    if [ "$OS" != "windows" ]; then
        chmod +x "${INSTALL_DIR}/${TARGET_BINARY}"
    fi

    print_success "Installed to ${INSTALL_DIR}/${TARGET_BINARY}"
}

while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            if [ -n "${2:-}" ]; then
                requested_version="$2"
                requested_version="${requested_version#v}"
                shift 2
            else
                print_error "--version requires a version argument"
                exit 1
            fi
            ;;
        --no-modify-path)
            no_modify_path=true
            shift
            ;;
        *)
            print_warning "Unknown option '$1'"
            shift
            ;;
    esac
done

main() {
    print_logo

    if [ -n "$requested_version" ]; then
        VERSION="$requested_version"
        print_info "Installing version: ${BOLD}v${VERSION}${NC}"
    else
        printf "  ${MUTED}Fetching latest version...${NC}\r"
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            echo ""
            print_error "Could not determine latest version"
            exit 1
        fi
        printf "                                    \r"
        print_success "Latest version: ${BOLD}v${VERSION}${NC}"
    fi

    existing_version=$(check_existing_version)
    if [ -n "$existing_version" ]; then
        clean_existing=$(echo "$existing_version" | sed 's/^v//')
        clean_version=$(echo "$VERSION" | sed 's/^v//')
        if [ "$clean_existing" = "$clean_version" ]; then
            print_info "Version v${VERSION} already installed"
            exit 0
        else
            print_info "Upgrading from v${clean_existing} to v${clean_version}"
        fi
    fi

    OS=$(get_os)
    ARCH=$(detect_rosetta)

    echo ""
    download_and_install "$VERSION" "$OS" "$ARCH"

    configure_path

    if [ -n "${GITHUB_ACTIONS-}" ] && [ "${GITHUB_ACTIONS}" = "true" ]; then
        echo "$INSTALL_DIR" >> "$GITHUB_PATH"
        print_success "Added to \$GITHUB_PATH"
    fi

    echo ""
    print_info "Run ${BOLD}remove-comments --help${NC} to get started"
    print_info "Repo: ${MUTED}https://github.com/${REPO}${NC}"
    echo ""
}

main
