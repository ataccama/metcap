#!/bin/sh

export PATH=/bin:/sbin:/usr/bin:/usr/sbin
echo -n "Adding metcap user and group... "
useradd -r -d /etc/metcap -s /bin/false -U metcap
echo "done."
