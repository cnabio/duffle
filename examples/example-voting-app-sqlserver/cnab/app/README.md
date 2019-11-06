Example Voting App
==================

**This app is from the following repo: https://github.com/twright-msft/example-voting-app**

Details on root application:
This is an example Docker app with multiple services. It is run with Docker Compose and uses Docker Networking to connect containers together. You will need Docker Compose 1.6 or later.
More info at https://blog.docker.com/2015/11/docker-toolbox-compose/

Details on this fork:
This fork replaces PostgreSQL with SQL Server as the backend database engine.
The Java worker application now uses the JDBC driver from Microsoft.
https://msdn.microsoft.com/en-us/library/mt484311(v=sql.110).aspx

The Node.js application now uses the open source tedious driver.
http://pekim.github.io/tedious/

Currently this version has a SQL Server image from the SQL Server on Linux private preview in the docker-compose.yml file.
In order to access that image, you need to be a participant in the SQL Server on Linux private preview.  
Sign up here: http://sqlserveronlinux.com
Once you have access to the image, make sure that you run 'docker login' and login once to cache your credentials before you run 'docker-compose up'.

Currently, this version uses the new JSON query features of SQL Server 2016.
https://msdn.microsoft.com/en-us/library/dn921897.aspx
There will be a new SQL Server 2016 Developer Edition container image published to Docker Hub soon!


Architecture
-----

* A Python webapp which lets you vote between two options
* A Redis queue which collects new votes
* A Java worker which consumes votes and stores them in…
* A SQL Server
* A Node.js webapp which shows the results of the voting in real time

Running
-------

Run in this directory:

    $ docker-compose up

The voting web app will be available on port 5000 on your Docker host, and the results web app will be on port 5001.
