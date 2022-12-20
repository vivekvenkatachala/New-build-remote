#! usr/bin/bash
dotnet sonarscanner begin /k:".net_coree" /d:sonar.host.url="http://20.106.156.255:9000"  /d:sonar.login="fea2829b213d06c013bf27bda3d7aa2efabd1ee1"
dotnet build
dotnet sonarscanner end /d:sonar.login="fea2829b213d06c013bf27bda3d7aa2efabd1ee1"