#!/bin/sh

set -eu

AP_VERSIONS=$(curl -sH"Accept: application/vnd.github.v3+json" https://api.github.com/repos/solo-io/autopilot/releases | python -c "import sys; from json import loads as l; releases = l(sys.stdin.read()); print('\n'.join(release['tag_name'] for release in releases))")

if [ "$(uname -s)" = "Darwin" ]; then
  OS=darwin
else
  OS=linux
fi

for AP_VERSION in $AP_VERSIONS; do

  tmp=$(mktemp -d /tmp/autopilot.XXXXXX)
  filename="ap-${OS}-amd64"
  url="https://github.com/solo-io/autopilot/releases/download/${AP_VERSION}/${filename}"

  if curl -f ${url} >/dev/null 2>&1; then
    echo "Attempting to download Autopilot CLI version ${AP_VERSION}"
  else
    continue
  fi

  (
    cd "$tmp"

    echo "Downloading ${filename}..."

    SHA=$(curl -sL "${url}.sha256" | cut -d' ' -f1)
    curl -sLO "${url}"
    echo "Download complete!, validating checksum..."
    checksum=$(openssl dgst -sha256 "${filename}" | awk '{ print $2 }')
    if [ "$checksum" != "$SHA" ]; then
      echo "Checksum validation failed." >&2
      exit 1
    fi
    echo "Checksum valid."
  )

  (
    cd "$HOME"
    mkdir -p ".autopilot/bin"
    mv "${tmp}/${filename}" ".autopilot/bin/ap"
    chmod +x ".autopilot/bin/ap"
  )

  rm -r "$tmp"

  echo "Autopilot CLI was successfully installed 🎉"
  echo ""
  echo "Add the Autopilot CLI CLI to your path with:"
  echo "  export PATH=\$HOME/.autopilot/bin:\$PATH"
  echo ""
  echo "Now run:"
  echo "  ap init myproject     # generate a new project directory"
  echo "Please see visit the Autopilot guides for more:  https://docs.solo.io/autopilot/latest"
  exit 0
done

echo "No versions of ap found."
exit 1
