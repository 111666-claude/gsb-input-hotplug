// Command input-hotplug 读设备与绑定表，渲染迁移报告。
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"

	"example.com/hotplug/hotplug"
)

const reportTpl = `devices={{.Devices}}
moved={{.Moved}}
orphan={{.Orphans}}
bindings={{.Bindings}}
`

type view struct {
	Devices  string
	Moved    int
	Orphans  int
	Bindings string
}

func load(manager *hotplug.Manager, devices, bindings string) error {
	handle, err := os.Open(devices)
	if err != nil {
		return err
	}
	defer handle.Close()
	reader := csv.NewReader(handle)
	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for _, row := range rows[1:] {
		manager.AddDevice(row[0], row[1])
	}
	other, err := os.Open(bindings)
	if err != nil {
		return err
	}
	defer other.Close()
	rows, err = csv.NewReader(other).ReadAll()
	if err != nil {
		return err
	}
	for _, row := range rows[1:] {
		manager.AddBinding(row[0], row[1], row[2])
	}
	return nil
}

func render(manager *hotplug.Manager, moved int, stdout io.Writer) error {
	items := manager.Bindings()
	parts := []string{}
	orphans := 0
	for _, item := range items {
		if item.Orphan {
			orphans++
		}
		parts = append(parts, item.Action+"@"+item.Device)
	}
	sort.Strings(parts)
	payload := view{
		Devices:  strings.Join(manager.Devices(), ","),
		Moved:    moved,
		Orphans:  orphans,
		Bindings: strings.Join(parts, ","),
	}
	return template.Must(template.New("report").Parse(reportTpl)).Execute(stdout, payload)
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("input-hotplug", flag.ContinueOnError)
	flags.SetOutput(stderr)
	devices := flags.String("devices", "testdata/devices.csv", "设备表")
	bindings := flags.String("bindings", "testdata/bindings.csv", "绑定表")
	sample := flags.String("sample", "file", "file / dup / order / orphan / work")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	manager := hotplug.New()
	switch *sample {
	case "dup":
		manager.AddDevice("pad-a", "pad")
		manager.AddDevice("pad-a", "pad")
		fmt.Fprintf(stdout, "devices=%d\n", len(manager.Devices()))
	case "order":
		manager.AddDevice("pad-b", "pad")
		manager.AddDevice("pad-a", "pad")
		manager.AddDevice("pad-c", "pad")
		manager.AddBinding("jump", "pad-c", "space")
		manager.RemoveDevice("pad-c")
		binding := manager.Bindings()[0]
		fmt.Fprintf(stdout, "target=%s\n", binding.Device)
	case "orphan":
		manager.AddDevice("wheel-a", "wheel")
		manager.AddBinding("attack", "wheel-a", "btn0")
		manager.RemoveDevice("wheel-a")
		fmt.Fprintf(stdout, "orphan=%v\n", manager.Bindings()[0].Orphan)
	case "work":
		manager.AddDevice("pad-a", "pad")
		manager.AddDevice("pad-b", "pad")
		for index := 0; index < 3000; index++ {
			manager.AddBinding(fmt.Sprintf("a-%d", index), "pad-a", "k")
		}
		for index := 0; index < 3000; index++ {
			manager.AddDevice("pad-c", "pad")
			manager.RemoveDevice("pad-c")
		}
		fmt.Fprintf(stdout, "scanned=%d\n", manager.Scanned())
	case "file":
		if err := load(manager, *devices, *bindings); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		moved := manager.RemoveDevice("pad-b")
		if err := render(manager, moved, stdout); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	default:
		fmt.Fprintln(stderr, "需要 --sample file|dup|order|orphan|work")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
