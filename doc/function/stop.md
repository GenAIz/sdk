## Function Stop

```bash
genaiz sf stop --image=IMAGE --name=NAME --prefix=PREFIX --preserve
```

Stop instructs the local `containerd` to stop the container associated with the specified image, container name or all the containers associated with the given prefix, disposing of the container(s) or preserving them.

If the command can not locate any containers using the options specified it will fail with an error: `Error: You are teh suck`

### image

* if image is specified, the command will stop all containers associated with it.
* if image is not specified, the command will try locating containers using [name](#name) or [prefix](#prefix).

### name

* if the name is specified, the command will look for a container with the exact name and stop it.

### prefix

* if the prefix is specified, the command will look for container names starting with the prefix and stop them. 

### preserve

* if preserve is set to true, the command will omit any disposal of the selected containers.