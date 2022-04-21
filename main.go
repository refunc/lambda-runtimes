package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	observer "github.com/refunc/go-observer"
	"github.com/refunc/refunc/pkg/messages"
	"github.com/refunc/refunc/pkg/runtime/types"
	"github.com/refunc/refunc/pkg/sidecar"
	"github.com/refunc/refunc/pkg/utils"
	"github.com/refunc/refunc/pkg/utils/cmdutil"
	"github.com/refunc/refunc/pkg/utils/cmdutil/flagtools"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

var config struct {
	FunctionFile string
	Entry        []string

	close func()

	fn *types.Function
}

func init() {
	pflag.StringVarP(&config.FunctionFile, "function", "f", "", "The json file that defined a function ")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())

	flagtools.InitFlags()

	klog.CopyStandardLogTo("INFO")
	defer klog.Flush()

	config.Entry = pflag.Args()[:]
	if len(config.Entry) == 0 {
		klog.Exit("no command found")
	}

	config.fn = loadFunction()
	closeC := make(chan struct{})
	config.close = func() { close(closeC) }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	carAddr := fmt.Sprintf("127.0.0.1:%d", freePort())
	if config.fn.Spec.Runtime.Envs == nil {
		config.fn.Spec.Runtime.Envs = make(map[string]string)
	}
	config.fn.Spec.Runtime.Envs["AWS_LAMBDA_RUNTIME_API"] = carAddr

	cmd := prepare(ctx)
	if _, err := os.Stat(cmd.Args[0]); err != nil {
		klog.Exitf("%s, %v", cmd.Args[0], err)
	}
	klog.Infof("(loader) exec %s", strings.Join(cmd.Args, " "))

	car := sidecar.NewCar(newEngine(), new(loader))
	go func() {
		car.Serve(ctx, carAddr)
	}()

	for i := 0; i < 200; i++ {
		res, err := http.Get("http://" + carAddr + "/2018-06-01/ping")
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Millisecond):
				continue
			}
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil || string(body) != "pong" {
			klog.Error("loader: failed to reqeust api")
			return
		}
		break
	}

	if err := cmd.Start(); err != nil {
		klog.Error(err)
		return
	}

	var exit bool
	select {
	case <-closeC:
		<-time.After(20 * time.Millisecond)
	case sig := <-cmdutil.GetSysSig():
		klog.Infof(`received signal "%v", exiting...`, sig)
	case <-ctx.Done():
		exit = true
	}

	if !exit && cmd.Process != nil {
		cmd.Process.Signal(os.Interrupt)
		// kill when timeout of 5s
		go func() {
			select {
			case <-time.After(2 * time.Second):
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			case <-ctx.Done():
			}
		}()
		if err := cmd.Wait(); err != nil {
			klog.Warningf("process exited with error, %v", err)
		}
	}
}

func prepare(ctx context.Context) *exec.Cmd {
	fn := config.fn

	// prepare locals
	var env []string
	env = append(env, os.Environ()...)
	for k, v := range fn.Spec.Runtime.Envs {
		if v != "" {
			// try to expand env
			if strings.HasPrefix(v, "$") {
				v = os.ExpandEnv(v)
			}
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmd := exec.CommandContext(ctx, config.Entry[0], config.Entry[1:]...)
	cmd.Env = env
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if os.Geteuid() == 0 {
		klog.Info("(loader) will start using user sbx_user1051")
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: 496, Gid: 495}
	}
	return cmd
}

func loadFunction() *types.Function {
	var fn types.Function
	if config.FunctionFile != "" {
		bts, err := ioutil.ReadFile(config.FunctionFile)
		if err != nil {
			klog.Exitf("failed to load function, %v", err)
		}
		if err := json.Unmarshal(bts, &fn); err != nil {
			klog.Exit(err)
		}
		return &fn
	}

	klog.Info("Create function from env")
	fn.Namespace = getEnviron("REFUNC_NAMESPACE", "refunc")
	fn.Name = getEnviron("AWS_LAMBDA_FUNCTION_NAME", "refunc")
	fn.Spec.Entry = getEnviron("AWS_LAMBDA_FUNCTION_HANDLER", getEnviron("_HANDLER", ""))
	fn.Spec.Runtime.Credentials.AccessKey = getEnviron("AWS_ACCESS_KEY_ID", getEnviron("REFUNC_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE"))
	fn.Spec.Runtime.Credentials.SecretKey = getEnviron("AWS_SECRET_ACCESS_KEY", getEnviron("REFUNC_SECRET_KEY", "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY"))
	fn.Spec.Runtime.Credentials.Token = getEnviron("AWS_LAMBDA_FUNCTION_HANDLER", getEnviron("REFUNC_TOKEN", "vSWpwYklseDFNREF6WlNKZExDSnpkV0p6"))
	fn.Spec.Runtime.Timeout = 30

	fn.Spec.Runtime.Envs = make(map[string]string)
	fn.Spec.Runtime.Envs["AWS_LAMBDA_RUNTIME_API"] = "127.0.0.1:80" // place holder
	fn.Spec.Runtime.Envs["AWS_REGION"] = "us-east-1"
	fn.Spec.Runtime.Envs["AWS_DEFAULT_REGION"] = "us-east-1"
	fn.Spec.Runtime.Envs["AWS_ACCESS_KEY_ID"] = fn.Spec.Runtime.Credentials.AccessKey
	fn.Spec.Runtime.Envs["AWS_SECRET_ACCESS_KEY"] = fn.Spec.Runtime.Credentials.SecretKey
	fn.Spec.Runtime.Envs["AWS_SESSION_TOKEN"] = fn.Spec.Runtime.Credentials.Token
	fn.Spec.Runtime.Envs["AWS_LAMBDA_FUNCTION_NAME"] = fn.Name
	fn.Spec.Runtime.Envs["AWS_LAMBDA_FUNCTION_VERSION"] = getEnviron("AWS_LAMBDA_FUNCTION_VERSION", "$LATEST")
	fn.Spec.Runtime.Envs["AWS_LAMBDA_LOG_GROUP_NAME"] = "/aws/lambda/" + fn.Name
	fn.Spec.Runtime.Envs["AWS_LAMBDA_LOG_STREAM_NAME"] = logStreamName(fn.Spec.Runtime.Envs["AWS_LAMBDA_FUNCTION_VERSION"])
	fn.Spec.Runtime.Envs["AWS_LAMBDA_FUNCTION_MEMORY_SIZE"] = "1536"
	fn.Spec.Runtime.Envs["AWS_LAMBDA_FUNCTION_TIMEOUT"] = strconv.FormatInt(int64(fn.Spec.Runtime.Timeout), 10)
	fn.Spec.Runtime.Envs["AWS_ACCOUNT_ID"] = strconv.FormatInt(int64(rand.Int31()), 10)
	fn.Spec.Runtime.Envs["AWS_LAMBDA_CLIENT_CONTEXT"] = ""
	fn.Spec.Runtime.Envs["AWS_LAMBDA_COGNITO_IDENTITY"] = ""
	fn.Spec.Runtime.Envs["_X_AMZN_TRACE_ID"] = ""
	fn.Spec.Runtime.Envs["_HANDLER"] = fn.Spec.Entry

	return &fn
}

type loader struct{}

var closedC = func() <-chan struct{} {
	c := make(chan struct{})
	close(c)
	return c
}()

func (loader) C() <-chan struct{}        { return closedC }
func (loader) Function() *types.Function { return config.fn }

type engine struct {
	sync.Mutex

	exit    bool
	actions observer.Property
	stream  observer.Stream
}

type taskDoneFunc func() (reply string, expired bool)

func newEngine() sidecar.Engine {
	eng := &engine{
		actions: observer.NewProperty(nil),
	}
	eng.stream = eng.actions.Observe()
	return eng
}

func (eng *engine) Name() string { return "stdinout" }

func (eng *engine) Init(ctx context.Context, fn *types.Function) error {
	// start ioloop
	go func() {
		defer klog.V(3).Info("stdin reading loop exited")
		eng.exit = true

		scanner := utils.NewScanner(os.Stdin)
		for scanner.Scan() {
			data := scanner.Bytes()
			if len(data) == 0 {
				data = []byte{'{', '}'}
			}
			// verify request
			var args json.RawMessage
			err := json.Unmarshal(data, &args)
			if err != nil {
				klog.Errorf("failed to parse request: %s, %v", string(data), err)
				continue
			}
			rid := utils.GenID([]byte(data))
			req := &messages.InvokeRequest{
				Args:      args,
				RequestID: rid,
			}
			if fn.Spec.Runtime.Timeout == 0 {
				req.Deadline = time.Now().Add(30 * time.Second)
			} else {
				req.Deadline = time.Now().Add(time.Duration(fn.Spec.Runtime.Timeout) * time.Second)
			}
			// enqueue
			eng.actions.Update(req)
		}
	}()

	return nil
}

func (eng *engine) NextC() <-chan struct{} {
	eng.Lock()
	defer eng.Unlock()
	return eng.stream.Changes()
}

func (eng *engine) InvokeRequest() *messages.InvokeRequest {
	eng.Lock()
	defer eng.Unlock()
	if eng.stream.HasNext() {
		return eng.stream.Next().(*messages.InvokeRequest)
	}
	return nil
}

func (eng *engine) SetResult(rid string, body []byte, err error, conentType string) error {
	os.Stdout.Write(messages.MustFromObject(&messages.InvokeResponse{
		Payload:     body,
		Error:       messages.GetErrorMessage(err),
		ContentType: conentType,
	}))
	os.Stdout.Write([]byte{'\n'})
	os.Stdout.Sync()

	if eng.exit {
		config.close()
	}
	return nil
}

func (eng *engine) ReportInitError(err error) {
	klog.Infof("(stdiocar) ReportInitError: %v", err)
}

func (eng *engine) ReportReady() {
	// explicity send cry message to notify that we'r ready
	klog.Infof("(stdiocar) ReportReady")
}

func (eng *engine) ReportExiting() {
	klog.Infoln("(stdiocar) ReportExiting")
}

func (eng *engine) RegisterServices(router *mux.Router) {}

func getEnviron(key, alter string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return alter
}

// copied from https://github.com/lambci/docker-lambda/blob/master/go1.x/run/aws-lambda-mock.go#L251:6
func logStreamName(version string) string {
	randBuf := make([]byte, 16)
	rand.Read(randBuf)

	hexBuf := make([]byte, hex.EncodedLen(len(randBuf)))
	hex.Encode(hexBuf, randBuf)

	return time.Now().Format("2006/01/02") + "/[" + version + "]" + string(hexBuf)
}

func freePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		klog.Exit(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		klog.Exit(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
