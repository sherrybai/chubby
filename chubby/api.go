package chubby

import (
	"cos518project/raft"
)

type API interface {
	RaftLeader() raft.ServerAddress
}

type APIImpl struct {

}

//// Return the address of the current Raft leader.
//func (api *APIImpl) RaftLeader() raft.ServerAddress {
//
//}
//
//// Mode type. Describes how a file handle will be used.
//type Mode int
//const (
//	ModeRead Mode = iota  // For reading
//	ModeWriteLock         // For writing and locking
//	ModeACL               // For changing the ACL
//)
//
//// File handle type.
//type FileHandle int
//
//// Open a named file or directory to produce a file handle.
//func Open(fileName string, mode Mode, createNew bool) FileHandle {
//
//}
//
//// Destroy a file handle. This call never fails.
//func Close(fileHandle FileHandle) {
//
//}
//
//// Cause outstanding and subsequent operations on this file handle to fail
//// without closing the file handle.
//func Poison(fileHandle FileHandle) {
//
//}
//
//// Return both the contents and the metadata of the file associated with this
//// file handle.
//func GetContentsAndStat(fileHandle FileHandle) {
//
//}
//
//// Return the metadata of the file associated with this file handle.
//func GetStat(fileHandle FileHandle) {
//
//}
//
//// Write the contents of the file associated with this file handle.
//// TODO: file contents param?
//func SetContents(fileHandle FileHandle) {
//
//}
//
//// Write the ACL names associated with the node associated with this file
//// handle.
//// TODO: 'names' should be list of strings?
//func SetACL(fileHandle FileHandle) {
//
//}
//
//// Delete the node associated with this file handle if the node has no
//// children.
//func Delete(fileHandle FileHandle) {
//
//}
//
//// Acquire a lock associated with this file handle. Block if the lock is
//// currently held by another client.
//func Acquire(fileHandle FileHandle) {
//
//}
//
//// Try to acquire a lock associated with this file handle. Return True on
//// success, False otherwise. This function is non-blocking.
//func TryAcquire(fileHandle FileHandle) bool {
//
//}
//
//// Release the lock held by this file handle.
//func Release(fileHandle FileHandle) {
//
//}
//
//// Return a sequencer that describes any lock held by this handle.
//// TODO: define sequencer type
//func GetSequencer() {
//
//}
//
//// Associate a sequencer with a handle. Subsequent operations on the handle
//// fail if the sequencer is no longer valid.
//func SetSequencer(fileHandle FileHandle) {
//
//}
//
//// Return if the sequencer is valid.
//func CheckSequencer() bool {
//
//}