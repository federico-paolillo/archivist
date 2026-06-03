#!/usr/bin/env sh

set -e
set -u

if [ "$#" -ne 5 ]; then
    echo "Usage: $0 <version> <gateway-image> <worker-image> <ui-image> <snapshotter-image>" >&2
    exit 2
fi

version="$1"
gateway_image="$2"
worker_image="$3"
ui_image="$4"
snapshotter_image="$5"

package_dir="release/compose"
package_file="release/archivist-compose-${version}.tar.gz"

rm -rf release
mkdir -p "${package_dir}"

cp docker-compose.prod.yaml "${package_dir}/docker-compose.yml"
cp .env.example "${package_dir}/.env"
cp rp.Caddyfile "${package_dir}/"

{
    printf 'ARCHIVIST_GATEWAY_IMAGE=%s\n' "${gateway_image}"
    printf 'ARCHIVIST_WORKER_IMAGE=%s\n' "${worker_image}"
    printf 'ARCHIVIST_UI_IMAGE=%s\n' "${ui_image}"
    printf 'ARCHIVIST_SNAPSHOTTER_IMAGE=%s\n' "${snapshotter_image}"
} > "${package_dir}/.env.images"

tar -czf "${package_file}" -C "${package_dir}" .
tar -tzf "${package_file}"
