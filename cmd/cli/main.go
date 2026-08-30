package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "time/tzdata"

	ze "github.com/securjty/zeopoxa-exporter"
)

var (
	Version = ""
)

func cmd() int {

	var (
		dsn         string
		dir         string
		timeout     time.Duration
		timezone    string
		showVersion bool
		logLevelRaw string
	)
	flag.StringVar(&dsn, "d", "", "Path to zeopoxa backup sqlite file")
	flag.StringVar(&dir, "o", "./exported_gpx", "Output dir for exported gpx files")
	flag.DurationVar(&timeout, "t", 90*time.Second, "Timeout parse db")
	flag.StringVar(&timezone, "z", "Europe/Moscow", "Timezone for parsing start track time")
	flag.StringVar(&logLevelRaw, "l", "info", "Log level. Available values: error, info, warn, debug")
	flag.BoolVar(&showVersion, "v", false, "Show version")

	flag.Parse()

	if showVersion {
		fmt.Println(Version)
		return 0
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevelRaw)); err != nil {
		level = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(level)

	if dsn == "" {
		slog.Error("empty database path")
		flag.Usage()
		return -1
	}

	zone, err := time.LoadLocation(timezone)
	if err != nil {
		slog.Error("invalid timezone", slog.Any("err", err))
		flag.Usage()
		return -1
	}

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		slog.Error("create output dir", slog.Any("err", err))
		flag.Usage()
		return -1
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		slog.Error("open database", slog.Any("err", err))
		return -1
	}

	err = db.Ping()
	if err != nil {
		slog.Error("ping database", slog.Any("err", err))
		return -1
	}
	slog.Info("parsing tracks from zeopoxa database")
	data, err := ze.Parse(ctx, db, zone)
	if err != nil {
		slog.Error("parse", slog.Any("err", err))
		return -1
	}
	slog.Info("parsing was successful")

	slog.Info("export to gpx files", slog.String("directory", dir))
	err = ze.Export(dir, data)
	if err != nil {
		slog.Error("export", slog.Any("err", err))
		return -1
	}
	slog.Info("done")

	return 0
}

func main() {
	os.Exit(cmd())
}
