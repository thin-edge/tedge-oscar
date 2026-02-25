## Flow packaging

A flow package is a tarball file (optionally compressed with gzip) which contains the following files listed in the table:

|Name|Description|Required|
|----|-----------|--------|
| flow.toml | Flow definition which contains the flow's input and output definitions, along with a list of the steps to be executed on the input. | Yes |
| *.js | JavaScript files referenced from the flow.toml file. Typically called "main.js", but can be called anything as the flow.toml contains the reference to the files | No, but typically included | 
| params.toml.template | Template file to provide guidance for users to create their own `params.toml` file which use used to control any exposed parameterized values used with the flow. | No |
| params.toml | Optional User defined parameters used to User-overridable config/parameter file which is never overwritten when flows are updated. | No |

Below shows an example of the contents of a simple flow package:

```sh
$ tar tvf /thingsboard:0.0.1.tar
-rw-r--r--  0 0      0        4343 Jan  1  1970 dist/main.js
-rw-r--r--  0 0      0         493 Jan  1  1970 flow.toml
```

**flow.toml**

The flow.toml is what makes a flow a flow. Without it, the flow's engine will not load the flow. It defines what inputs should be fed into the flow, what SmartFunctions should be executed, and also where the output messages should be sent to.

The steps typically reference JavaScript (`*.js`) files which are also included in the flow, but may also reference built-in functions provided by the flows engine.

**params.toml**

The `params.toml` is an optional file, that can be created by the user to customize specific aspects of the deployed flow without modifying the original flow definition (`flow.toml`) or JavaScript code. The `params.toml` file is persisted across flow updates to ensure that any user-parameterized values are not lost if the flow is updated. The `flow.toml` file should reference the values using the template syntax. This enables flow authors to clearly communicate which values should be editable for users, and which values should not.

A flow can provide an example file called `params.toml.template` which contains the list of parameterizable values that the user can set. To parameterize a flow, the user should copy the `params.toml.template` file to the `params.toml` location, and then customize any of the values inside the file. Ideally the `params.toml.template` includes some documentation about each parameter to assist the user.

Below shows a simplistic example of the parameterization of a flow:

**file: flow.toml (created by the flow)**

```toml
# meta information about the flow
name = "foo"
version = 1.0.0

input.mqtt.topics = ["foo"]

[[steps]]
script = "main.js"
params.debug = "${.params.debug}"
```

**file: main.js (created by the flow)**

```js
onMessage(message, context) {
    if (context.params.debug) {
        // Print message to help with debugging
        console.log("Received message", {message});
    }
    return [{
        topic: "te/device/main///e/foo",
        payload: JSON.stringify({
            text: `Received a message on ${message.topic}`,
        }),
    }];
}
```

**file: params.toml.template (created by the flow)**
```toml
# enable debugging
debug = false
```

**file: params.toml (created by the user)**

The params.toml file can be created by first copying the contents 

```sh
# copy 
cp params.toml.template cp params.toml
```

Then edit the values accordingly:

```toml
# enable debugging
debug = true
```

### Example flow deployments

Below shows the directory structure where multiple flows are installed under the Cumulocity mapper, and the "local" mapper (i.e. the generic flows engine).

**Directory: /etc/tedge/mappers/**

```
/etc/tedge/mappers
├── c8y
│   ├── events
│   │   ├── flow.toml
│   │   └── main.js
│   └── measurements
│       ├── flow.toml
│       └── main.js
└── local
    ├── thingsboard
    │   ├── commands
    │   │   ├── flow.toml
    │   │   └── main.js
    │   └── telemetry
    │       ├── flow.toml
    │       └── main.js
    └── uptime
        ├── params.toml
        ├── params.toml.template
        ├── flow.toml
        └── main.js
```

## Flow Management

Flow are are to be managed by a thin-edge.io software management plugin, with the following steps:

1. Install
1. Update
1. Remove

An example of each of the commands is detailed in the following sections.

### List

The list command lists all of the flows from all of the mappers. A prefix is used to show a flows deployment to a specific mapper in the form of:

```sh
{mapper}/{name}
```

* The `{mapper}` value is determined from the mapper name in which the flow is located
* The name is determined from the `.name` property of the `flow.toml`
* The version is determined from the `.version` property of the `flow.toml`

Below shows an example of the list command showing the installed flows in the corresponding mappers.

```sh
# list
$ ./flows list

c8y/certificate-alert   1.0.0
```

### Install/Update

A flow is installed by unpacking the contents of the tarball under the designated mapper's flows directory.

Below shows examples of the sm-plugins used to install a new flow.

```sh
# install (to 'flows' mapper)
./flows install local/thingsboard --module-version 0.0.1 --file /opt/homebrew/etc/tedge/flows/images/thingsboard:0.0.1.tar
./flows install local/certificate-alert --module-version 1.0.0 --file /opt/homebrew/etc/tedge/flows/images/certificate-alert:1.0.0.tar
```

**Notes**

* Any existing "params.toml" file should be preserved and not overwritten.

### Remove

A flow is deleted using the following steps:

1. Delete all files and folder under the flow's directory except the "params.toml" file (if present). If a "params.toml" file is not present, then the flow's directory can also be deleted.

Below shows examples of a flow is removed using the software management plugin:

```sh
# remove (version is optional)
./flows remove local/certificate-alert --module-version "1.0.0"
./flows remove local/thingsboard
```



### Required changes

* [ ] Support relative address lookup based on the flow toml definition file rather than the mapper relative address - https://github.com/thin-edge/thin-edge.io/issues/3947
* [ ] Only listen for flow definitions which are called "flow.toml" instead of all .toml files
* [ ] Support loading an optional "params.toml" and expose the values as "${.params}" when evaluating the flow, at least in the topics:
    * topics
    * params (previously named `config`)


### Questions

* How to declare a flow's dependency to built-in flow engine functions like, `add-timestmap`?

### Requirements

* Allow users to parameterize instances of flows on a device which are not deleted if the flow is updated remotely
