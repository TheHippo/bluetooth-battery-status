# bluetooth-battery-status

A commandline tool to show the battery status of Bluetooth devices that can be found through `bluetoothctl`.

## Installation

    go install github.com/TheHippo/bluetooth-battery-status

## Example

    $ bluetooth-battery-status
    +-------------------+--------------------------+-----------+---------+
    | MAC               | NAME                     | CONNECTED | BATTERY |
    +-------------------+--------------------------+-----------+---------+
    | 68:6C:E6:80:C5:5F | Xbox Wireless Controller | true      |     100 |
    +-------------------+--------------------------+-----------+---------+