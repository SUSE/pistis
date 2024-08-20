#
# spec file for package pistis
#
# Copyright (c) 2024 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#


Name:           pistis
Version:        0
Release:        0
Summary:        Git commit verification
License:        GPL-3.0-or-later
URL:            https://github.com/SUSE/pistis
Source0:        %{name}-%{version}.tar.gz
Source1:        vendor.tar.gz
BuildRequires:  go >= 1.21
#BuildRequires:  golang-packaging

%description
Tool for verifying the commits in a Git repository.

%prep
%autosetup -a1 -p1

%build
go build -buildmode=pie -mod=vendor -v -x

%install
install -d %{buildroot}%{_bindir}
install %{name} %{buildroot}%{_bindir}

# test vendoring seems not working / "cannot find module ..."
#check
#gotest .

%files
%license COPYING
%doc README.md
%{_bindir}/%{name}

%changelog
