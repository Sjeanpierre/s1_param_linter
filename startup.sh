#!/usr/bin/env bash

export TUNNEL_ORIGIN_CERT="$(pwd)/cf_warp_cert.pem"
openssl enc -d -aes-256-cbc -in cf_warp_cert.pem.enc -out cf_warp_cert.pem -k $WARP_CERT_DECRYPT_KEY
chmod +x ./cloudflare-warp
nohup ./cloudflare-warp --hostname $WARP_DOMAIN http://localhost:4545 >> tunnel.log 2>&1 & echo $!
nohup ./gh_enforcer >> app.log 2>&1 & echo $!
tail -qF *.log