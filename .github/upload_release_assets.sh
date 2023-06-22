#!/usr/bin/env bash
set -e

main() {
  check_params "$@"
  upload_assets
}

check_params() {
  if [[ -z "${GITHUB_TOKEN}" ]]
  then
    echo "Error: environent variable GITHUB_TOKEN must be set"
    exit 1
  fi

  if [[ -z "$2" ]]
  then
    echo "Usage: $0 <directory with release assets> <release upload URL>"
    exit 1
  else
    asset_path="$1"
    upload_url="$2"
  fi
}

upload_assets() {
  url="$(echo $upload_url | sed 's/{?name,label}//')"
  echo "Uploading assets to ${url}..."
  for file_path in "${asset_path}"/*
  do
    file_name="$(basename ${file_path})"
    echo "Uploading ${file_path}..."
    curl --fail -L \
        -X POST \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${GITHUB_TOKEN}"\
        -H "X-GitHub-Api-Version: 2022-11-28" \
        -H "Content-Type: application/octet-stream" \
        "${url}?name=${file_name}" \
        --data-binary "@${file_path}"
    printf "\n\n"
  done
}

main "$@"
