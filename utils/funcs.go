package utils

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

// HmacSHA1buf .
func HmacSHA1buf(key string, data string) []byte {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func HasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func GetCurrentPath() (string, error) {
	path, err := os.Executable()
	if nil != err {
		return "", err
	}
	exePath := filepath.Dir(path)
	return exePath, err
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CheckFileIsExist(filename string) bool {
	finfo, err1 := os.Stat(filename)
	if os.IsNotExist(err1) {
		return false
	}
	return finfo != nil && err1 == nil
}

// partitionString partitions the string into chunks of given size,
// with the last chunk of variable size.
func PartitionString(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		panic("invalid chunkSize")
	}
	length := len(s)
	chunks := 1 + length/chunkSize
	start := 0
	end := chunkSize
	parts := make([]string, 0, chunks)
	for {
		if end > length {
			end = length
		}
		parts = append(parts, s[start:end])
		if end == length {
			break
		}
		start, end = end, end+chunkSize
	}
	return parts
}

func SaveJson(destFile string, obj interface{}) error {
	file, err := os.OpenFile(destFile, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		file, err = os.Create(destFile)
		if nil != err {
			Logger.Warn(fmt.Sprintf("Open file failed Err:%s", err.Error()),
				zap.String("file", destFile))
			return err
		}
	}
	defer file.Close()
	buf, err := json.MarshalIndent(obj, "", "\t")
	if nil != err {
		Logger.Error("save json false!",
			zap.String("err", err.Error()),
			zap.String("file", destFile))
		return err
	}
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(string(buf))
	if nil != err {
		Logger.Error("write json false!",
			zap.String("err", err.Error()),
			zap.String("file", destFile))
		return err
	}
	writer.Flush()
	return nil
}

func StopPidGroup(pid int32, destFolder string) {
	proc, err := process.NewProcess(pid)
	if nil != err {
		Logger.Error("stop false",
			zap.String("err", err.Error()),
			zap.Int32("pid", pid))
		return
	}
	pidMpa := map[int32]bool{}
	StopProcGroup(proc, destFolder, &pidMpa)
}

func GetPidGroup(pid int32, pids *[]int32) {
	if pid > 0 {
		proc, err := process.NewProcess(pid)
		if nil == err {
			*pids = append(*pids, proc.Pid)
			chs, err := proc.Children()
			if nil == err {
				for _, ch := range chs {
					GetPidGroup(ch.Pid, pids)
				}
			}
		}
	}
}

func StopProcGroup(dest *process.Process, destFolder string, pidMpa *map[int32]bool) {
	if nil != dest {

		if _, ok := (*pidMpa)[dest.Pid]; ok {
			return
		}

		(*pidMpa)[dest.Pid] = true
		cmdline, _ := dest.Cmdline()
		if strings.Contains(cmdline, destFolder) {
			child, err := dest.Children()
			if nil == err {
				for _, ch := range child {
					StopProcGroup(ch, destFolder, pidMpa)
				}
			}

			err = dest.Kill()

			if nil == err {
				if nil != Logger {
					Logger.Info("stop success",
						zap.String("cmdline", cmdline))
					return
				}
			} else {
				if nil != Logger {
					Logger.Error("stop false",
						zap.String("err", err.Error()),
						zap.String("cmdline", cmdline))
				}
			}
			if nil != Logger {
				Logger.Warn("stop false",
					zap.String("cmdline", cmdline))
			}
		}
	}
}

func CreateShortcut(dst, src, folder, desc string) error {
	ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY)
	oleShellObject, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer oleShellObject.Release()
	wshell, err := oleShellObject.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()
	cs, err := oleutil.CallMethod(wshell, "CreateShortcut", dst)
	if err != nil {
		return err
	}
	idispatch := cs.ToIDispatch()
	oleutil.PutProperty(idispatch, "TargetPath", src)
	oleutil.PutProperty(idispatch, "WorkingDirectory", folder)
	oleutil.PutProperty(idispatch, "Description ", desc)
	oleutil.CallMethod(idispatch, "Save")
	return nil
}

// machineId stores machine id generated once and used in subsequent calls
// to NewObjectId function.
var machineId = readMachineId()
var processId = os.Getpid()

// readMachineId generates and returns a machine id.
// If this function fails to get the hostname it will cause a runtime error.
func readMachineId() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		n := uint32(time.Now().UnixNano())
		sum[0] = byte(n >> 0)
		sum[1] = byte(n >> 8)
		sum[2] = byte(n >> 16)
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

// readRandomUint32 returns a random objectIdCounter.
func readRandomUint32() uint32 {
	// We've found systems hanging in this function due to lack of entropy.
	// The randomness of these bytes is just preventing nearby clashes, so
	// just look at the time.
	return uint32(time.Now().UnixNano())
}

var objectIdCounter uint32 = readRandomUint32()

// NewObjectId returns a new unique ObjectId.
func NewObject() []byte {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	b[7] = byte(processId >> 8)
	b[8] = byte(processId)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return b[:]
}

func OnTimeConsuming(beg int64, name string) {
	dif := time.Now().UnixMilli() - beg
	if dif > 100 {
		Logger.Warn(name, zap.Int64("time consuming", dif))
	} else {
		Logger.Info(name, zap.Int64("time consuming", dif))
	}
}

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func Int64ToBytes(n int64) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func BytesToInt64(bys []byte) int64 {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}
