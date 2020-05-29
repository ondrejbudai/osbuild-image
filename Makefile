#
# Maintenance Helpers
#
# This makefile contains targets used for development, as well as helpers to
# aid automatization of maintenance. Unless a target is documented in
# `make help`, it is not supported and is only meant to be used by developers
# to aid their daily development work.
#
# All supported targets honor the `SRCDIR` variable to find the source-tree.
# For most unsupported targets, you are expected to have the source-tree as
# your working directory. To specify a different source-tree, simply override
# the variable via `SRCDIR=<path>` on the commandline. By default, the working
# directory is used for build output, but `BUILDDIR=<path>` allows overriding
# it.
#

BUILDDIR ?= .
SRCDIR ?= .

#
# Automatic Variables
#
# This section contains a bunch of automatic variables used all over the place.
# They mostly try to fetch information from the repository sources to avoid
# hard-coding them in this makefile.
#
# Most of the variables here are pre-fetched so they will only ever be
# evaluated once. This, however, means they are always executed regardless of
# which target is run.
#
#     VERSION:
#         This evaluates the `Version` field of the specfile. Therefore, it will
#         be set to the latest version number of this repository without any
#         prefix (just a plain number).
#
#     COMMIT:
#         This evaluates to the latest git commit sha. This will not work if
#         the source is not a git checkout. Hence, this variable is not
#         pre-fetched but evaluated at time of use.
#

VERSION := $(shell (cd "$(SRCDIR)" && grep "^Version:" osbuild-image.spec | sed 's/[^[:digit:]]*\([[:digit:]]\+\).*/\1/'))
COMMIT = $(shell (cd "$(SRCDIR)" && git rev-parse HEAD))

#
# Maintenance Targets
#
# The following targets are meant for development and repository maintenance.
# They are not supported nor is their use recommended in scripts.
#

.PHONY: build
build:
	go build -o osbuild-image ./cmd/osbuild-image/

.PHONY: install
install:
	cp osbuild-image /usr/bin/

#
# Building packages
#
# The following rules build osbuild-image packages from the current HEAD
# commit, based on the spec file in this directory.  The resulting packages
# have the commit hash in their version, so that they don't get overwritten
# when calling `make rpm` again after switching to another branch.
#
# All resulting files (spec files, source rpms, rpms) are written into
# ./rpmbuild, using rpmbuild's usual directory structure.
#

RPM_SPECFILE=rpmbuild/SPECS/osbuild-image-$(COMMIT).spec
RPM_TARBALL=rpmbuild/SOURCES/osbuild-image-$(COMMIT).tar.gz

$(RPM_SPECFILE):
	mkdir -p $(CURDIR)/rpmbuild/SPECS
	(echo "%global commit $(COMMIT)"; git show HEAD:osbuild-image.spec) > $(RPM_SPECFILE)

$(RPM_TARBALL):
	mkdir -p $(CURDIR)/rpmbuild/SOURCES
	git archive --prefix=osbuild-image-$(COMMIT)/ --format=tar.gz HEAD > $(RPM_TARBALL)

.PHONY: srpm
srpm: $(RPM_SPECFILE) $(RPM_TARBALL)
	rpmbuild -bs \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--with tests \
		$(RPM_SPECFILE)

.PHONY: rpm
rpm: $(RPM_SPECFILE) $(RPM_TARBALL)
	rpmbuild -bb \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--with tests \
		$(RPM_SPECFILE)
