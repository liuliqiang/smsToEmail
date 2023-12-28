%define debug_package %{nil}

Name:    %{_name}
Version: %{_version}
Release: %{_release}%{?dist}
Summary: %{name}
License: MIT
URL:     https://liqiang.io

Source0: %{_source}.tar.gz
Source1: %{name}.service
Source2: %{name}.env

%description
%{name}


%prep
%setup -q -n %{_source}


%install
install -d -m 755 %{buildroot}%{_bindir}
install -c -m 755 %{name} %{buildroot}%{_bindir}/%{name}

install -d -m 755 %{buildroot}%{_unitdir}
install -c -m 644 %{SOURCE1} %{buildroot}%{_unitdir}/%{name}.service

install -d -m 755 %{buildroot}%{_sysconfdir}/sysconfig
install -c -m 644 %{SOURCE2} %{buildroot}%{_sysconfdir}/sysconfig/%{name}

%post

%preun
if [ $1 -eq 0 ]; then
  # uninstall
  /bin/systemctl disable %{name}.service
  /bin/systemctl stop %{name}.service
fi


%files
%defattr(-,root,root,-)
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%config(noreplace) %{_sysconfdir}/sysconfig/%{name}

%changelog

