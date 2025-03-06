#!/bin/bash

#set -x

ROOT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)

generatePlugins() {
  plugins=(
    $(tanzu plugin group get vmware-tanzu/platform-engineer -o json | \
        jq -r '.[] | "\(.pluginname):\(.plugintarget):\(.pluginversion)"')
  )

  for info in "${plugins[@]}"; do
    name=$(echo $info | cut -d: -f1)
    target=$(echo $info | cut -d: -f2)
    version=$(echo $info | cut -d: -f3)

    for os in darwin windows linux; do
      for arch in amd64 arm64; do
          mkdir -p $ROOT_DIR/plugins/vmware/publisher/$os/$arch/$target
          plugin=$ROOT_DIR/plugins/vmware/publisher/$os/$arch/$target/$name:$version
          cat << EOFplugin > $plugin
#!/bin/bash

# Copyright 2025 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Minimally Viable Dummy Tanzu CLI 'Plugin'

info() {
   cat << EOF
{
  "name": "$name",
  "target": "$target",
  "description": "$name functionality",
  "version": "$version",
  "buildSHA": "01234567",
  "group": "System",
  "hidden": false,
  "aliases": [],
  "completionType": 0
}
EOF
  exit 0
}

case "$1" in
    info)  $1 "$@";;
    env) env;;
    *) cat << EOF
Stub plugin "$name" "$version" for "$target"
EOF
       ;;
esac
EOFplugin
      done
    done
  done
}

generateInventory() {
  imgpkg pull -o $ROOT_DIR/plugins/tmp -i projects.packages.broadcom.com/tanzu_cli/plugins/plugin-inventory:latest
  mv $ROOT_DIR/plugins/tmp/* $ROOT_DIR/plugins
  rmdir $ROOT_DIR/plugins/tmp
}

generatePlugins
generateInventory