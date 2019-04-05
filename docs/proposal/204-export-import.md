# Duffle Export & Import

This document covers how to export and import a bundle.

## Export
Consider the case where a user wants to create package that contains the bundle manifest along with the all of the necessary artifacts to execute the install/uninstall/lifecycle actions specified in the invocation images. You can use the `$ duffle export [BUNDLE_REFERENCE]` command to do just that.

Duffle `export` allows a user to create a gzipped tar archive that contains the bundle manifest along with all of the necessary images including the invocation images and the referenced images in the bundle. See example below

### Export Example
```console
$ duffle bundle list
NAME      	VERSION	DIGEST
helloworld	0.1.1  	92145d4132aba06e11940a79f20402a3621196f1
wordpress 	0.2.0  	b91550cfc20bd21929e48b21c88715c9e89349eb

$ duffle export wordpress:0.2.0
$ ls
wordpress-0.2.0.tgz
```

In the example, you'll find the exported artifact, a gzipped tar archive: `wordpress-0.2.0.tgz`. Unpacking that artifact results in the following directory structure:
```
wordpress-0.2.0/
   bundle.cnab
   artifacts/
      cnab-wordpress-0.2.0.tar
```

Duffle export gives users the ability to package up all the components of their distributed application along with the logic to manage that application leaving users with a portable artifact.

## Import

Duffle import is used to import the exported artifact above along with all of the necessary images to manage the application. It unpacks the artifact and saves all of the images in the `artifacts/` to the local Docker store.

### Import Example
```console
$ duffle import wordpress-0.2.0.tgz

$ ls
wordpress-0.2.0.tgz wordpress-0.2.0/

$ ls wordpress-0.2.0/
bundle.json artifacts/

$ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
cnab/wordpress      latest              533cfdfba95a        5 hours ago         35MB
```

## Future Improvements
- Handle non-docker images (issue [#513](https://github.com/deislabs/duffle/issues/513), [#494](https://github.com/deislabs/duffle/issues/494))
- Handle exporting and importing from duffle's local bundle store (issue [#379](https://github.com/deislabs/duffle/issues/379))
