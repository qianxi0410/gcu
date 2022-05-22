package main

import "errors"

var errCanNotFindGoModFile = errors.New("can't find go.mod file in your designated path")
