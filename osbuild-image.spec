%global goipath         github.com/ondrejbudai/osbuild-image

Version:        1

%gometa

%global common_description %{expand:
A simple CLI frontend for osbuild-composer focused solely on building images
from a blueprint.
}

Name:           osbuild-image
Release:        1%{?dist}
Summary:        A simple CLI frontend for osbuild-composer

# osbuild-composer doesn't have support for building i686 images
# and also RHEL and Fedora has now only limited support for this arch.
ExcludeArch:    i686

# Upstream license specification: Apache-2.0
License:        ASL 2.0
URL:            %{gourl}
Source0:        %{gosource}


BuildRequires:  %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}
%if 0%{?fedora}
#BuildRequires:  golang(github.com/aws/aws-sdk-go)
#BuildRequires:  golang(github.com/Azure/azure-sdk-for-go)
#BuildRequires:  golang(github.com/Azure/azure-storage-blob-go/azblob)
#BuildRequires:  golang(github.com/BurntSushi/toml)
#BuildRequires:  golang(github.com/coreos/go-semver/semver)
#BuildRequires:  golang(github.com/coreos/go-systemd/activation)
#BuildRequires:  golang(github.com/google/uuid)
#BuildRequires:  golang(github.com/julienschmidt/httprouter)
#BuildRequires:  golang(github.com/gobwas/glob)
#BuildRequires:  golang(github.com/google/go-cmp/cmp)
#BuildRequires:  golang(github.com/stretchr/testify/assert)
%endif

Requires: osbuild-composer >= 13

%description
%{common_description}

%prep
%if 0%{?rhel}
%forgeautosetup -p1
%else
%goprep
%endif

%build
%if 0%{?rhel}
GO_BUILD_PATH=$PWD/_build
install -m 0755 -vd $(dirname $GO_BUILD_PATH/src/%{goipath})
ln -fs $PWD $GO_BUILD_PATH/src/%{goipath}
cd $GO_BUILD_PATH/src/%{goipath}
install -m 0755 -vd _bin
export PATH=$PWD/_bin${PATH:+:$PATH}
export GOPATH=$GO_BUILD_PATH:%{gopath}
export GOFLAGS=-mod=vendor
%endif

%gobuild -o %{gobuilddir}/bin/osbuild-image %{goipath}/cmd/osbuild-image

%install
install -m 0755 -vd                     %{buildroot}%{_bindir}
install -m 0755 -vp %{gobuilddir}/bin/* %{buildroot}%{_bindir}/

%check
%if 0%{?rhel}
export GOFLAGS=-mod=vendor
export GOPATH=$PWD/_build:%{gopath}
%gotest ./...
%else
%gocheck
%endif

%files
%license LICENSE
%doc README.md
%{_bindir}/*

%changelog
# the changelog is distribution-specific, therefore it doesn't make sense to have it upstream
