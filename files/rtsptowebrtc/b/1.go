package main

import (
    "log"
    "os/exec"
)

func main() {

    cmd := exec.Command("setzoomdown")

    err := cmd.Run()

    if err != nil {
        log.Fatal(err)
    }
}
