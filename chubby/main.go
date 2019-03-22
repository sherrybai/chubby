// Adapted from Leto command interface:
// https://github.com/yongman/leto/blob/master/cmd/main.go

// Copyright (C) 2018 YanMing <yming0221@gmail.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package chubby

import (
	"cos518project/chubby/config"
	"cos518project/chubby/server"
	"flag"
)

var (
	listen		string
	raftDir 	string
	raftBind 	string
	nodeId		string
	join		string
	inmem		bool
)

func init() {
	flag.StringVar(&listen, "listen", ":5379", "server listen port")
	flag.StringVar(&raftDir, "raftdir", "./", "raft data directory")
	flag.StringVar(&raftBind, "raftbind", ":15379", "raft bus transport bind port")
	flag.StringVar(&nodeId, "id", "", "node id")
	flag.StringVar(&join, "join", "", "join to existing cluster at this address")
	flag.BoolVar(&inmem, "inmem", false, "log and stable storage in memory")
}

func main() {
	// Parse flags from command line.
	flag.Parse()

	var (
		c *config.Config
	)

	// Create new Chubby config.
	c = config.NewConfig(listen, raftDir, raftBind, nodeId, join, inmem)

	// Create new app.
	app := server.NewApp(c)

	// Run the app.
	go app.Run()
}
