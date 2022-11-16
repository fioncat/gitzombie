#!/bin/bash

set -eu

targets=( \
	"darwin_amd64" \
	"darwin_arm64" \
	"linux_amd64" \
	"linux_arm64" \
)

BOLD="$(tput bold 2>/dev/null || printf '')"
GREY="$(tput setaf 0 2>/dev/null || printf '')"
UNDERLINE="$(tput smul 2>/dev/null || printf '')"
RED="$(tput setaf 1 2>/dev/null || printf '')"
GREEN="$(tput setaf 2 2>/dev/null || printf '')"
YELLOW="$(tput setaf 3 2>/dev/null || printf '')"
BLUE="$(tput setaf 4 2>/dev/null || printf '')"
MAGENTA="$(tput setaf 5 2>/dev/null || printf '')"
CYAN="$(tput setaf 6 2>/dev/null || printf '')"
RESET="$(tput sgr0 2>/dev/null || printf '')"

info() {
	printf '%s\n' "${BOLD}${GREY}>${RESET} ${CYAN}$*${RESET}"
}

error() {
	printf '%s\n' "${RED}x $*${RESET}" >&2
}

shell_join() {
	local arg
	printf "%s" "$1"
	shift
	for arg in "$@"; do
		printf " "
		printf "%s" "${arg// /\ }"
	done
}

confirm() {
	read -p "$1 (y/n) " -n 1 -r
	echo
	if [[ $REPLY =~ ^[Yy]$ ]]; then
		return 0
	fi
	error "user aborted"
	exit 1
}

execute() {
	shell_exec=$(shell_join "$@")
	if ! "$@"; then
		error "failed to execute command"
		exit 1
	fi
}

has() {
	command -v "$1" 1>/dev/null 2>&1
}

download() {
	file="$1"
	url="$2"

	if has curl; then
		execute "curl" "--fail" "--location" "--output" "$file" "$url"
	elif has wget; then
		execute "wget" "--output-document=$file" "$url"
	elif has fetch; then
		execute "fetch" "--output=$file" "$url"
	else
		error "No HTTP download program (curl, wget, fetch) found, exiting…"
		return 1
	fi
}

# Test if a location is writeable by trying to write to it.
test_writeable() {
	path="${1:-}/test.txt"
	if touch "${path}" 2>/dev/null; then
		rm "${path}"
		return 0
	else
		return 1
	fi
}

# Currently supporting:
#   - amd64
#   - arm64
detect_arch() {
	arch="$(uname -m | tr '[:upper:]' '[:lower:]')"
	case "${arch}" in
		amd64|x86_64) arch="amd64" ;;
		arm64) arch="arm64" ;;
	esac
	printf '%s' "${arch}"
}

ensure_command() {
	if has $1; then
		return 0
	fi
	error "command $1 is required to install gitzombie"
}

ensure_command "perl"
ensure_command "unzip"

BIN_DIR="/usr/local/bin"
TMP_DIR="/tmp/gitzombie-install"
BASE_URL="https://github.com/fioncat/gitzombie/releases"

PLATFORM="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(detect_arch)"

TARGET="${PLATFORM}_${ARCH}"
URL="${BASE_URL}/latest/download/gitzombie-${TARGET}.zip"

SUPPORT=""
for support_target in "${targets[@]}"; do
	if [[ "${TARGET}" == "${support_target}" ]]; then
		SUPPORT="true"
	fi
done

if [ -z ${SUPPORT} ]; then
	error "Sorry, now we donot support your platform: ${TARGET}"
	exit 1
fi

confirm "Install gitzombie to ${BIN_DIR}?"

if [ -d ${TMP_DIR} ]; then
	rm -r ${TMP_DIR}
fi
mkdir -p ${TMP_DIR}
ARCHIVE_FILE="${TMP_DIR}/gitzombie.zip"
info "Downloading gitzombie"
download ${ARCHIVE_FILE} ${URL}

info "Unzipping file"
execute "unzip" "-qq" "${TMP_DIR}/gitzombie.zip" -d "${TMP_DIR}/out"

TMP_BIN_PATH="${TMP_DIR}/out/bin/gitzombie"
if test_writeable "${BIN_DIR}"; then
	info "Moving binary file"
	execute "mv" "${TMP_BIN_PATH}" "${BIN_DIR}"
else
	info "Escalated permissions are required to install to ${BIN_DIR}"
	execute "sudo" "mv" "${TMP_BIN_PATH}" "${BIN_DIR}"
fi

rm -r ${TMP_DIR}
cat << EOF

Congratulations, gitzombie has been already installed to ${CYAN}${BIN_DIR}${RESET}.
You should add the init script to your shell profile according to the
instruction below.

For bash user, add this to your ~/.bashrc:
   ${GREEN}source <(gitzombie init bash)${RESET}

For zsh user, add this to your ~/.zshrc:
   ${GREEN}source <(gitzombie init zsh)${RESET}

For more details, please refer to: https://github.com/fioncat/gitzombie
EOF
