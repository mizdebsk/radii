[![build status](https://img.shields.io/github/actions/workflow/status/mizdebsk/radii/ci.yml?branch=main)](https://github.com/mizdebsk/radii/actions/workflows/ci.yml?query=branch%3Amain)
[![License](https://img.shields.io/github/license/mizdebsk/radii.svg?label=License)](https://www.gnu.org/licenses/gpl-3.0-standalone.html)


radii
=====

Hardware driver manager for RHEL

The radii tool is a command-line utility that provides a
consistent interface for installing and maintaining third-party
AI-accelerator and GPU driver stacks on Red Hat Enterprise Linux
(RHEL).  It detects compatible accelerator hardware, enables the
appropriate package repositories, and installs the required kernel and
user-mode components from RHEL-distributed packages.  The tool manages
multi-component driver stacks end-to-end and integrates with the
standard dnf workflow so that installed drivers are updated through
normal system package management.

The name "radii" comes from RDI (RHEL Drivers Installation), the core
purpose of the tool.  It echoes the acronym while also being a real
word.  Radii is the plural of the Latin radius, which originally meant
a ray, rod, or spoke.  The plural form reflects the many pieces that
make up a driver stack.  In anatomy, the radii are the bones that let
your hands move, serving as a simple metaphor for enabling hardware to
work.

Copying
-------

This is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free
Software Foundation, either version 3 of the License, or (at your
option) any later version.

This software is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
General Public License for more details.

A copy of the GNU General Public License is contained in the COPYING
file.
