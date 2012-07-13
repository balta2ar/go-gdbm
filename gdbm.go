// Packge gdbm implements a wrapper around libgdbm for Go.
package gdbm

// #cgo CFLAGS: -std=gnu99
// #cgo LDFLAGS: -lgdbm
// #include <stdlib.h>
// #include <gdbm.h>
// #include <string.h>
// inline datum mk_datum(char * s) {
//     datum d;
//     d.dptr = s;
//     d.dsize = strlen(s);
//     return d;
// }
import "C"

import (
    "errors"
    "unsafe"
)

//
type Database struct {
    dbf C.GDBM_FILE
    mode int
}

//
type DatabaseCfg struct {
    Mode string
    BlockSize int
    Permissions int
}

func lastError() error {
    return errors.New(C.GoString(C.gdbm_strerror(C.gdbm_errno)))
}

// Simple function to open a database file with default parameters (block size
// is default for the filesystem and file permissions are set to 0666).
func Open(filename string, mode string) (db * Database, err error) {
    return OpenWithCfg(filename, DatabaseCfg{mode, 0, 0666})
}

func OpenWithCfg(filename string, cfg DatabaseCfg) (db * Database, err error) {
    db = new(Database)
    err = nil

    switch cfg.Mode {
    case "r": db.mode = C.GDBM_READER
    case "w": db.mode = C.GDBM_WRITER
    case "c": db.mode = C.GDBM_WRCREAT
    case "n": db.mode = C.GDBM_NEWDB
    }

    cs := C.CString(filename)
    defer C.free(unsafe.Pointer(cs))

    db.dbf = C.gdbm_open(cs, C.int(cfg.BlockSize), C.int(db.mode), C.int(cfg.Permissions), nil)
    if db.dbf == nil {
        err = lastError()
    }
    return db, err
}

// Closes a database file.
func (db * Database) Close() {
    C.gdbm_close(db.dbf)
}

//
func (db * Database) update(key string, value string, flag C.int) (err error) {
    kcs := C.CString(key)
    vcs := C.CString(value)
    k := C.mk_datum(kcs)
    v := C.mk_datum(vcs)
    defer C.free(unsafe.Pointer(kcs))
    defer C.free(unsafe.Pointer(vcs))

    retv := C.gdbm_store(db.dbf, k, v, flag)
    if retv != 0 {
        err = lastError()
    }
    return err
}

//
func (db * Database) Insert(key string, value string) (err error) {
    return db.update(key, value, C.GDBM_INSERT)
}

//
func (db * Database) Replace(key string, value string) (err error) {
    return db.update(key, value, C.GDBM_REPLACE)
}

func (db * Database) Exists(key string) bool {
    kcs := C.CString(key)
    k := C.mk_datum(kcs)
    defer C.free(unsafe.Pointer(kcs))

    e := C.gdbm_exists(db.dbf, k)
    if e == 1 {
        return true
    }
    return false
}

//
func (db * Database) Fetch(key string) (value string, err error) {
    kcs := C.CString(key)
    k := C.mk_datum(kcs)
    defer C.free(unsafe.Pointer(kcs))

    vdatum := C.gdbm_fetch(db.dbf, k)
    if vdatum.dptr == nil {
        return "", lastError()
    }

    value = C.GoString(vdatum.dptr)
    defer C.free(unsafe.Pointer(vdatum.dptr))
    return value, nil
}

//
func (db * Database) Delete(key string) (err error) {
    kcs := C.CString(key)
    k := C.mk_datum(kcs)
    defer C.free(unsafe.Pointer(kcs))

    retv := C.gdbm_delete(db.dbf, k)
    if retv == -1 && db.mode == C.GDBM_READER {
        err = lastError()
    }
    return err
}

//
func (db * Database) Reorganize() {
    C.gdbm_reorganize(db.dbf)
}

//
func (db * Database) Sync() {
    C.gdbm_sync(db.dbf)
}
