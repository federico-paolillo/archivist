#!/usr/bin/env sh

set -e
set -u

if [ "$#" -ne 15 ]; then
    echo "Usage: $0 <version> <target-ref> <release-sha> <gateway-image> <worker-image> <ui-image> <snapshotter-image> <gateway-digest> <worker-digest> <ui-digest> <snapshotter-digest> <gateway-attestation-url> <worker-attestation-url> <ui-attestation-url> <snapshotter-attestation-url>" >&2
    exit 2
fi

version="$1"
target_ref="$2"
release_sha="$3"
gateway_image="$4"
worker_image="$5"
ui_image="$6"
snapshotter_image="$7"
gateway_digest="$8"
worker_digest="$9"
ui_digest="${10}"
snapshotter_digest="${11}"
gateway_attestation_url="${12}"
worker_attestation_url="${13}"
ui_attestation_url="${14}"
snapshotter_attestation_url="${15}"

git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

git tag -a "${version}" "${release_sha}" -m "Release ${version}"
git push origin "refs/tags/${version}"

notes_file="$(mktemp)"
trap 'rm -f "${notes_file}"' EXIT

cat > "${notes_file}" <<EOF
Draft release for ${version}.

Target ref: ${target_ref}
Resolved commit: ${release_sha}

Images:

- [${gateway_image}@${gateway_digest}](${gateway_image}@${gateway_digest})
- [${worker_image}@${worker_digest}](${worker_image}@${worker_digest})
- [${ui_image}@${ui_digest}](${ui_image}@${ui_digest})
- [${snapshotter_image}@${snapshotter_digest}](${snapshotter_image}@${snapshotter_digest})

Attestations:

- [Gateway](${gateway_attestation_url})
- [Worker](${worker_attestation_url})
- [UI](${ui_attestation_url})
- [Snapshotter](${snapshotter_attestation_url})

Deployment package: archivist-compose-${version}.tar.gz, attached as a draft release asset.

Note: unknown/unknown OS/Arch entries are [Attestation Manifest Descriptors](https://docs.docker.com/build/metadata/attestations/attestation-storage/#attestation-manifest-descriptor).

Edit these notes before publishing the release.
EOF

gh release create "${version}" \
    --draft \
    --title "${version}" \
    --notes-file "${notes_file}"
