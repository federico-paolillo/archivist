#!/usr/bin/env sh

set -e
set -u

if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <version> <gateway-image> <worker-image> <ui-image>" >&2
    exit 2
fi

version="$1"
gateway_image="$2"
worker_image="$3"
ui_image="$4"

package_dir="release/compose"
package_file="release/archivist-compose-${version}.tar.gz"

rm -rf release
mkdir -p "${package_dir}"

cp docker-compose.prod.yaml "${package_dir}/"
cp docker-compose.prod.env.example "${package_dir}/"
cp rp.Caddyfile "${package_dir}/"

{
    printf 'ARCHIVIST_GATEWAY_IMAGE=%s\n' "${gateway_image}"
    printf 'ARCHIVIST_WORKER_IMAGE=%s\n' "${worker_image}"
    printf 'ARCHIVIST_UI_IMAGE=%s\n' "${ui_image}"
} > "${package_dir}/docker-compose.images.env"

tar -czf "${package_file}" -C "${package_dir}" .
tar -tzf "${package_file}"
