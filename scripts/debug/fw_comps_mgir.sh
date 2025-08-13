#!/bin/bash
BDF=${BDF:-${1}}
FW_COMPS_DEBUG=1 MFT_DEBUG=1 sudo -E mstflint -d $BDF query full | awk '/MGIR/,{p=1} p; /MCQI/{exit}'
