#!/bin/bash

cd /app/web/build
python3 -m http.server 23538 &
/app/celeve 