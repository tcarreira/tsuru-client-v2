#!/usr/bin/env sh
set -eu

git_authors="$(git shortlog -se)"
git_authors="$(echo "${git_authors}" | sed -E "s/[[:space:]]+/ /g" | cut -d' ' -f3- | sort)"
contributors=$(echo "${git_authors}" \
  | awk 'BEGIN { FS="<"}
    {arr[$1] = arr[$1] "<" $2 " "}
    END {for (i in arr) print i arr[i]}' \
  | sed 's/[[:space:]]$//' \
  | sort)

echo "# AUTOGENERATED FILE - DO NOT EDIT
#
# This is the official list of people who have contributed code to tsuru-client.
# The AUTHORS file lists the copyright holders; this file lists people.
# For example, Globo.com employees are listed here but not in AUTHORS,
# because Globo.com holds the copyright.
#
# This file was generated by misc/contributors_file.sh

${contributors}
"
