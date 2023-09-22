#!/bin/zsh
#

SHA="eb7c84570ae25672f0a9a9c11aab8471a2588fe5"
COMBINED=BenIsSuperAwesome$SHA
HANDSHAKE=$(echo -n "$COMBINED" | openssl sha1 | awk '{print $2}')

echo "SHA: $SHA"
echo "COMBINED: $COMBINED"
echo "HANDSHAKE: $HANDSHAKE"

curl -X PUT http://localhost:8080/deploy --verbose \
-H "Content-Type: application/json" \
-d "{ \"handshake\": \"$HANDSHAKE\", \"headsha\": \"$SHA\" }"


