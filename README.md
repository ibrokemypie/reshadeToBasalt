# README

A simple tool to convert Reshade presets into vkBasalt configurations.

## Prerequisites

``go>=1.12``

``git``

## Installation

Clone this repo then run ``go install cmd/reshadeToBasalt`` inside it.

To update simply pull then run the command again.

## Usage

``reshadeToBasalt -i <pathToReshadePreset.ini>``

This will attempt to generate a vkBasalt configuration based on the provided Reshade preset.

If this succeeds it should print the environment variables required to use it.

The generated vkBasalt configuration is not portable.

If output is not provided, it will output to a folder above the folder the preset ini is in, called ``<presetName>_vkBasalt``

## Flags

``-i    --input=    Path to the Reshade preset ini``

``-o    --output=   Path to output the vkBasalt configuration to    [optional]``

### Additionally

This generator was written with very minimal knowledge of how Reshade presets and vkBasalt works. It likely will not work for all cases, and there are likely many edge cases that will have to be fixed when run into (for example, "ContrastAdaptiveSharpen" is actually called "cas" and must be worked around).

Please open issues if you find other such cases.

This generator only works for Reshade presets, not SweetFX.
