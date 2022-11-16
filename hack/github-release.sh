#!/bin/bash

set -ue

targets=( \
	"darwin_amd64" \
	"darwin_arm64" \
	"linux_amd64" \
	"linux_arm64" \
)

VERSION=${GITHUB_REF#refs/*/}
echo "VERSION=${VERSION}" >> $GITHUB_ENV

make_asset() {
	release_os=$1
	release_arch=$2
	zip_file="gitzombie-${VERSION}-${release_os}_${release_arch}.zip"
	zip -r out/${zip_file} ./out/${release_os}_${release_arch}/bin LICENSE
}

mkdir -p out

for target in "${targets[@]}"; do
	echo "Build target: ${target}"
	IFS='_' read -r -a tmp <<< "$target"
	BUILD_OS="${tmp[0]}"
	BUILD_ARCH="${tmp[1]}"
	GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH}" go build -ldflags="-X 'main.Version=${VERSION}'" -o bin/gitzombie
	zip -r out/gitzombie-${VERSION}-${target}.zip ./bin LICENSE
done

