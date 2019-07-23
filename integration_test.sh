#! /bin/bash

set -euo pipefail

for file in xbase/testdata/*.encode.input; do
    echo "testing ${file}"
    diff <(/usr/bin/base64 "${file}") <(./build/base64 "${file}")
    diff <(/usr/bin/base64 --wrap=0 "${file}") <(./build/base64 --wrap=0 "${file}")
    diff <(/usr/bin/base64 -w 137 "${file}") <(./build/base64 -w 137 "${file}")
    diff <(/usr/bin/base64 -w 200 "${file}") <(./build/base64 -w 200 "${file}")
done

for file in xbase/testdata/*.decode.std.*.no-garbage.padded.input; do
    echo "testing ${file}"
    diff <(/usr/bin/base64 -d "${file}") <(./build/base64 -d "${file}")
    diff <(/usr/bin/base64 --decode "${file}") <(./build/base64 --decode "${file}")
done

for file in xbase/testdata/*.decode.std.*.std-garbage.padded.input; do
    echo "testing ${file}"
    diff <(/usr/bin/base64 -d -i "${file}") <(./build/base64 -d -i "${file}")
    diff <(/usr/bin/base64 --decode --ignore-garbage "${file}") <(./build/base64 --decode --ignore-garbage "${file}")
done
