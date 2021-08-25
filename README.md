# Companion App

### Requirements

- You will require Go `1.16` or higher.
- Java `11` or higher.
- `Makefile` support if you want to use the prebuilt build commands.
- Add a file called `service-account.json` into the root of `companion-cli` folder with valid GCP service account JSON credentials.

### How to build

Run the make file. The default rule will build the codebase for all required architectures. This will also build the two java apps for handle printer config and scale measurements.

```shell
$ make
```

The resulting binaries can be found inside the `bin` folder.

### Checksums

You can also optionally create a checksums file by running `make checksums`. This will output a `checksums.txt` file in the `bin` directory.

# How it works

The app is used to interact with printers and USB scales.
When the app starts it connects to Firestore to start listening for new in bound print jobs and scale jobs.

It also creates a small web server that Blade will attempt to ping in order to get the companion app reference for the locally running client.

This reference is passed as a header to all V5 & V6 API requests so the servers can add print jobs to the correct collection in firestore.

## Running
The program can be running manually by starting the executable in the terminal. No arguments are required. 

### Service 
It can also be installed to run on start up in the background as a service on linux, mac & windows.

#### Install
To install the app as a service run `companion_app --service install`
#### Uninstall
To uninstall the service run `companion_app --service uninstall`
#### Start
The service will auto start on login, however you can manually start the service if it is not running. Run `companion_app --service start`
#### Stop 
You can manually stop the service by running `companion_app --service start`
#### Restart 
You can stop and start the service by running `companion_app --service restart`
#### Status 
You can check if the service is running by calling `companion_app --service restart`

### Utils

There are also a few commands that can be used to help when debugging issues.

#### List Printers

You can list all the printers the app can see by running `companion_app --list-printers`

#### Print Test Page

You can print a sample PDF by running `companion_app --print-test-page=MY_PRINTER_NAME_HERE`

#### Read Scales

You can read the values from the connected USB scales by running `companion_app --read-scales`

# License 
This software is published under the AGPLv3 License. See OPEN_SOURCE for more information.
