# encoding: utf-8
# copyright: 2018, synamedia.com

title "Test prometheus RPM installation."

describe port(9100) do
  it { should be_listening }
end

describe file("/etc/sysconfig/prometheus/") do
  it { should be_directory }
  its("owner") { should eq "prometheus" }
end

describe file("/var/log/prometheus") do
  it { should be_directory }
  its("owner") { should eq "prometheus" }
end

describe file("/var/run/prometheus") do
  it { should be_directory }
  its("owner") { should eq "prometheus" }
end

describe file("/etc/init.d/prometheus-aggregate-exporter") do
  it { should be_file }
  its("owner") { should eq "prometheus" }
end

describe file("/etc/sysconfig/prometheus/prometheus-aggregate-exporter.env") do
  it { should be_file }
  its("owner") { should eq "prometheus" }
end

describe file("/var/log/prometheus/prometheus-aggregate-exporter.log") do
  it { should be_file }
  its("owner") { should eq "prometheus" }
end

describe file("/etc/sysconfig/prometheus/prometheus-aggregate-exporter.env") do
  its("content") { should match 'EXPORTER_ARGS\=\"-config=\/etc\/sysconfig\/prometheus\/prometheus-aggregate-exporter-config.yml\"' }
end

describe service("prometheus-aggregate-exporter") do
  it { should be_installed }
  it { should be_enabled }
  it { should be_running }
end

describe package("prometheus-aggregate-exporter") do
  it { should be_installed }
  its("version") { should eq "0.0.0-1" }
end

describe command("service prometheus-aggregate-exporter status") do
  its("stdout") { should match 'prometheus-aggregate-exporter \(pid  \d*\) is running...' }
  its("stderr") { should eq "" }
  its("exit_status") { should eq 0 }
end

describe command("service prometheus-aggregate-exporter stop") do
  its("stdout") { should eq "Stopping Prometheus prometheus-aggregate-exporter daemon: [  OK  ]\r\n" }
  its("stderr") { should eq "" }
  its("exit_status") { should eq 0 }
end

describe command("service prometheus-aggregate-exporter start") do
  its("stdout") { should eq "Starting Prometheus prometheus-aggregate-exporter daemon: [  OK  ]\r\n" }
  its("stderr") { should eq "" }
  its("exit_status") { should eq 0 }
end
