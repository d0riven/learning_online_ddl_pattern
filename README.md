## What's this?

For learning online ddl pattern on MySQL.  

### Requirement

* Docker
* golang compiler

### Usage

Task runner with Makefile.  
You move to each directory (copy_rename, replication, standard/{inplace,instant}) and input as follows via terminal:  

```
$ make setup
$ make test_online_ddl
```

### How it works

#### standard/{inplace, instant}

This directory is performed online ddl via mysql standard function.

#### replication

This is performed via M/M replication.

#### copy_rename

Performed via sequentially data copy, reflect changes to the current database by trigger each DML(INSERT/DELETE/UPDATE), and finally rename table.  
This method is also used by pt-online-schema-change , gh-ost, oak-online-alter-table, and so on.
