package zeopoxaexporter

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/tkrajina/gpxgo/gpx"
)

func Export(dir string, data []ZeopoxaTrack) error {
	const (
		op = "zeopoxaexporter.Export"

		KMH_2_MS           = 5.0 / 18.0
		CREATOR            = "zeopoxa-exporter"
		AUTHOR             = "zeopoxa-exporter"
		AUTHOR_LINK        = "github.com/securjty/zeopoxa-exporter"
		XMLNS              = "http://www.topografix.com/GPX/1/1"
		VERSION            = "1.1"
		XMLNS_XSI          = "http://www.w3.org/2001/XMLSchema-instance"
		XSI_SCHEMALOCATION = "http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd"

		EXPORT_ACTIVITY_TYPE = "cycling"
	)

	for _, item := range data {
		output := gpx.GPX{
			XMLNs:        XMLNS,
			XmlNsXsi:     XMLNS_XSI,
			XmlSchemaLoc: XSI_SCHEMALOCATION,
			Version:      VERSION,

			Creator:    CREATOR,
			AuthorName: AUTHOR,
			AuthorLink: AUTHOR_LINK,
		}
		points := make([]gpx.GPXPoint, 0, len(item.Points))
		output.RegisterNamespace("gpxtpx", "http://www.garmin.com/xmlschemas/TrackPointExtension/v3")
		for _, p := range item.Points {
			extension := gpx.ExtensionNode{}
			extension.XMLName = xml.Name{Space: "gpxtpx", Local: "TrackPointExtension"}

			extension.Nodes = append(extension.Nodes, gpx.ExtensionNode{
				XMLName: xml.Name{Space: "gpxtpx", Local: "speed"},
				Data:    strconv.FormatFloat(p.Speed*KMH_2_MS, 'f', 2, 64),
			})

			if p.HeartRate > 0 {
				extension.Nodes = append(extension.Nodes, gpx.ExtensionNode{
					XMLName: xml.Name{Space: "gpxtpx", Local: "hr"},
					Data:    strconv.FormatFloat(p.HeartRate, 'f', 0, 64),
				})
			}

			elevation := gpx.NewNullableFloat64(p.Elevation)
			points = append(points, gpx.GPXPoint{
				Timestamp:  p.Created.UTC(),
				Latitude:   p.Latidude,
				Longitude:  p.Longitude,
				Elevation:  *elevation,
				Extensions: gpx.Extension{Nodes: []gpx.ExtensionNode{extension}},
			})
		}

		track := gpx.GPXTrack{
			Name: "Zeopoxa Track",
			Type: EXPORT_ACTIVITY_TYPE,
			Segments: []gpx.GPXTrackSegment{
				{
					Points: points,
				},
			},
		}
		output.Tracks = append(output.Tracks, track)
		xmlData, err := output.ToXml(gpx.ToXmlParams{
			Version: VERSION,
			Indent:  true,
		})
		if err != nil {
			return fmt.Errorf("%s: convert to xml: %w", op, err)
		}
		name := fmt.Sprintf("zeopoxa_exporter_%d_%d.gpx", item.Id, item.StartTime.UnixMilli())
		path := filepath.Join(dir, name)

		f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("%s: open output gpx file: %w", op, err)
		}

		_, err = io.Copy(f, bytes.NewReader(xmlData))
		if err != nil {
			_ = f.Close()
			return fmt.Errorf("%s: write gpx to file: %w", op, err)
		}
		_ = f.Close()
	}

	return nil
}
