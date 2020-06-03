OSBuild Image
=============

A simple CLI frontend for osbuild-composer focused solely on building images
from a blueprint.

By default, it uploads a blueprint to osbuild-composer, starts a compose,
downloads it and deletes all artifacts from the osbuild-composer instance.

## Examples

* Build a minimal qcow2 image
  
  `osbuild-image --type qcow2 --output minimal.qcow2`
  
* Build an ami image from a blueprint (both json and toml are supported)
  
  `osbuild-image --type ami --output minimal.vhdx --blueprint bp.toml` 

* Build a minimal vhd image and save also the logs and manifest
  
  `osbuild-image --type vhd --output minimal.vhd --output-log log.txt --output-manifest manifest.json`

* Build a minimal qcow2 image and keep all artifacts inside osbuild-composer
  
  `osbuild-image --type vhd --output minimal.vhd --keep-artifacts`
