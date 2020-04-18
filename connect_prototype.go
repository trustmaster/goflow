// +build ignore

package goflow

import (
	"errors"
	"fmt"
	"reflect"
)

func connect(sendAddr, recvAddr portAddr) error {
	ch, err := attachPort(sendAddr, reflect.SendDir, nil)
	if err != nil {
		return err
	}
	_, err = attachPort(recvAddr, reflect.RecvDir, ch)
	return err
}

func attachPort(addr portAddr, dir reflect.ChanDir, ch reflect.Value) (reflect.Value, error) {
	proc := getProc(addr.proc)
	port := getPort(proc, addr.proc, addr.port)
	if addr.key != "" {
		return attachMapPort(port, addr.key, dir, ch)
	} else {
		return attachFieldPort(port, dir, ch)
	}
}

func attachMapPort(port reflect.Value, key string, dir reflect.ChanDir, ch reflect.Value) (reflect.Value, error) {
	err := validatePortType(port.Elem().Type())
	exch := port.MapIndex(key)
	ch, err = validChan(ch, exch)
	port.SetMapIndex(key, ch)
	return ch, nil
}

func attachFieldPort(port reflect.Value, dir reflect.ChanDir, ch reflect.Value) (reflect.Value, error) {
	err := validatePortType(port.Type(), dir)
	err = validateSettable(port)
	ch, err = validChan(ch, port)
	port.Set(ch)
	return ch, nil
}

func validChan(ch, exch reflect.Value) (reflect.Value, error) {
	if !exch.IsNil() {
		if ch.IsNil() {
			return exch, nil
		}
		return ch, errors.New("cannot attach new channel to already attached port")
	}
	if ch.IsNil() {
		ch = makeChan(ch.Type())
	}
	return ch, nil
}

func connectAddr() {
	// It can be an array port
	accessor := parsePortName(portName)
	if accessor.index >= 0 {
		// Expecting an array port
		if portType.Kind() == reflect.Slice {
			if portVal.Cap() == 0 {
				portVal = reflect.MakeSlice(portType, 0, 32)
			}
			if portVal.Len() <= accessor.index {
				portVal.SetLen(accessor.index + 1)
			}
			portType = portType.Elem()
			portVal = portVal.Index(accessor.index)
			if portVal.IsNil() {
				return nilValue, fmt.Errorf("Connect error: array port '%s.%s' has no index '%d'", procName, accessor.port, accessor.index)
			}
		} else {
			return nilValue, fmt.Errorf("Connect error: '%s.%s' is not an array port", procName, accessor.port)
		}
	} else if accessor.key != "" {
		// Expecting a hashmap port
		if portType.Kind() == reflect.Map {
			if portVal.IsZero() {
				portVal = reflect.MakeMap(portType)
			}
			portType = portType.Elem()
			portVal = portVal.MapIndex(reflect.ValueOf(accessor.key))
			// if portVal.IsNil() {
			// 	return nilValue, fmt.Errorf("Connect error: hashmap port '%s.%s' has no key '%s'", procName, accessor.port, accessor.key)
			// }
		} else {
			return nilValue, fmt.Errorf("Connect error: '%s.%s' is not an hashmap port", procName, accessor.port)
		}
	}
}
