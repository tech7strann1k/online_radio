package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/vansante/go-ffprobe"

	. "github.com/tech7strann1k/online-radio/db"
)

var click = 0
var timerCount = 0

type StreamPlayer struct {
	StreamTitle, StreamLogo, StreamUrl, playing_state string
	command                                           *exec.Cmd
}

func NewPlayer() *StreamPlayer {
	player := &StreamPlayer{}
	return player
}

func (player *StreamPlayer) Play() *StreamPlayer {
	player.playing_state = "playing"
	comm := exec.Command("ffplay", "-nodisp", "-i", player.StreamUrl)
	err := comm.Start()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("playing... %s\n", player.StreamUrl)
	player.command = comm
	return player
}

func (player *StreamPlayer) Stop() error {
	time.Sleep(time.Duration(time.Duration.Milliseconds(3)))
	_ = player.command.Process.Kill()
	return player.command.Wait()
}

func (player *StreamPlayer) ShowMetadata() (*ffprobe.StreamTags, error) {
	time.Sleep(time.Duration(time.Duration.Seconds(5)))
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeURL(ctx, player.StreamUrl)
	if err != nil {
		fmt.Println(err)
		ctx.Done()
		return nil, err
	}
	return &data.Streams[0].Tags, err
}

var builder, _ = gtk.BuilderNew()
var player = NewPlayer()
var db = InitDB("db/metadata.db")

type MainWindow struct {
	MainWindow   						*gtk.Window
	PlaylistView 						*gtk.ListBox
	PlayButton, StopButton, AddButton,
	PrefsButton, FavouritesButton 		*gtk.Button
	SelectCountryBox 					*gtk.ComboBoxText
	MetadataView     					*gtk.Label
	StreamLogoView   					*gtk.Image
	Player           					*StreamPlayer
}

type AddStreamDialog struct {
	AddStreamDialog                             		*gtk.Dialog
	AddStreamNameBox, AddStreamUrlBox, AddStreamIconBox	*gtk.Entry
	AddStreamIconButton, OkButton, CancelButton 		*gtk.Button
}

func NewMainWindow() *MainWindow {
	err := builder.AddFromFile("design.glade")
	if err != nil {
		fmt.Println(err)
	}
	obj, err := builder.GetObject("window_main")
	if err != nil {
		fmt.Println(err)
	}
	window := obj.(*gtk.Window)
	window.SetTitle("Online Radio")
	window.SetDefaultSize(720, 720)
	obj_2, err := builder.GetObject("playlist_widget")
	if err != nil {
		fmt.Println(err)
	}
	playlistView := obj_2.(*gtk.ListBox)
	obj_3, _ := builder.GetObject("play_button")
	playButton := obj_3.(*gtk.Button)
	obj_4, _ := builder.GetObject("stop_button")
	stopButton := obj_4.(*gtk.Button)
	obj_5, _ := builder.GetObject("stream_metadata_label")
	metadataView := obj_5.(*gtk.Label)
	obj_6, _ := builder.GetObject("stream_logo_view")
	streamLogoView := obj_6.(*gtk.Image)
	obj_7, _ := builder.GetObject("add_button")
	addButton := obj_7.(*gtk.Button)
	obj_8, _ := builder.GetObject("prefs_button")
	prefsButton := obj_8.(*gtk.Button)
	obj_9, _ := builder.GetObject("select_country_box")
	selectCountryBox := obj_9.(*gtk.ComboBoxText)
	obj_10, _ := builder.GetObject("favourites_button")
	favouritesButton := obj_10.(*gtk.Button)
	mainWindow := &MainWindow{MainWindow: window, PlaylistView: playlistView,
		PlayButton: playButton, StopButton: stopButton,
		MetadataView: metadataView, StreamLogoView: streamLogoView,
		AddButton: addButton, PrefsButton: prefsButton,
		SelectCountryBox: selectCountryBox, FavouritesButton: favouritesButton}
	return mainWindow
}

func (wnd *MainWindow) Init() {
	// LandList := LoadLandList()
	data := db.LoadData(nil)
	for _, elem := range data {
		row := addRow(elem)
		wnd.PlaylistView.Add(row)
	}
	wnd.PlaylistView.Connect("row-selected", func() {
		row := wnd.PlaylistView.GetSelectedRow()
		var (
			streamTitle string = data[row.GetIndex()].Title
			streamLogo         = data[row.GetIndex()].Logo
			streamUrl          = data[row.GetIndex()].Url
		)
		player.StreamTitle = streamTitle
		player.StreamLogo = streamLogo
		player.StreamUrl = streamUrl
		if click == 0 {
			player.playing_state = "started"
		} else {
			player.playing_state = "item_changed"
		}
		fmt.Println(player.playing_state)
		wnd.showMetadata()
	})
	d := time.Duration.Milliseconds(10)
	wnd.PlayButton.Connect("clicked", func() {
		click++
		if click == 1 {
			go player.Play()
		} else {
			go func() {
				m, _ := regexp.MatchString("stopped|item_changed", player.playing_state)
				if m {
					player.Stop()
					time.Sleep(time.Duration(d))
					player.playing_state = "playing"
					player.Play()
				}
			}()
		}
	})
	wnd.StopButton.Connect("clicked", func() {
		go func() {
			if player.playing_state != "stopped" {
				time.Sleep(time.Duration(d))
				player.playing_state = "stopped"
				player.Stop()
			}
		}()
	})
	wnd.AddButton.Connect("clicked", func() {
		item, _, err := dlgs.Entry("Add Stream", "", "http://stream.m3u")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(item)
	})
	window := wnd.MainWindow
	window.Connect("destroy", func() {
		if player.playing_state == "playing" {
			player.Stop()
		}
		gtk.MainQuit()
	})
	window.ShowAll()
}

func NewAddStreamDialog() *AddStreamDialog {
	err := builder.AddFromFile("design.glade")
	if err != nil {
		fmt.Println(err)
	}
	obj, err := builder.GetObject("add_stream_dialog")
	if err != nil {
		fmt.Println(err)
	}
	dialog := obj.(*gtk.Dialog)
	dialog.SetTitle("Add Stream")
	obj_2, _ := builder.GetObject("add_stream_name_box")
	addStreamNameBox := obj_2.(*gtk.Entry)
	obj_3, _ := builder.GetObject("add_stream_url_box")
	addStreamUrlBox := obj_3.(*gtk.Entry)
	obj_4, _ := builder.GetObject("add_stream_icon_box")
	addStreamIconBox := obj_4.(*gtk.Entry)
	obj_5, _  := builder.GetObject("add_stream_url_button")
	addStreamIconButton := obj_5.(*gtk.Button)
	obj_6, _ := builder.GetObject("ok_button")
	okButton := obj_6.(*gtk.Button)
	obj_7, _ := builder.GetObject("cancel_button")
	cancelButton := obj_7.(*gtk.Button)
	addStreamDialog := &AddStreamDialog{AddStreamDialog: dialog, AddStreamNameBox: addStreamNameBox, AddStreamUrlBox: addStreamUrlBox, 
									    AddStreamIconBox: addStreamIconBox, AddStreamIconButton: addStreamIconButton,
										OkButton: okButton, CancelButton: cancelButton}
	return addStreamDialog
}

func (dlg *AddStreamDialog) Init() {
	var initPath string
	dlg.AddStreamIconButton.Connect("clicked", func ()  {
		file, _, err := dlgs.File("Open file", "*.png *.jpg", true)
		if err != nil {
			fmt.Println(err)
		}
		initPath = file
		dlg.AddStreamIconBox.SetText(initPath)
	})
	dlg.OkButton.Connect("clicked", func ()  {
		var dirs = strings.Split(initPath, "/")
		var filename = dirs[len(dirs)]
		os.Chdir("radio_logos")
		destPath := fmt.Sprintf("./%s", filename)
		_, err := os.Stat(destPath)
		if os.IsNotExist(err) {
			initFile, _ := os.Open(initPath)
			defer initFile.Close()
			destFile, _ := os.Create(destPath)
			defer  destFile.Close()
			io.Copy(initFile, destFile)
		}
		streamName, _ := dlg.AddStreamNameBox.GetText()
		streamUrl, _ := dlg.AddStreamUrlBox.GetText()
		db.AddData(streamName, streamUrl, filename)
	})
}

func main() {
	gtk.Init(nil)
	wnd := NewMainWindow()
	wnd.Init()
	dlg := NewAddStreamDialog()
	dlg.Init()
	gtk.Main()
}

func addRow(metadata Stream) *gtk.ListBoxRow {
	hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 7)
	var streamLogo = fmt.Sprintf("radio_logos/%s", metadata.Logo)
	pixbuf, _ := gdk.PixbufNewFromFileAtScale(streamLogo, 32, 32, true)
	logoImage, _ := gtk.ImageNewFromPixbuf(pixbuf)
	logoLabel, _ := gtk.LabelNew(metadata.Title)
	hbox.Add(logoImage)
	hbox.Add(logoLabel)
	row, _ := gtk.ListBoxRowNew()
	row.Add(hbox)
	return row
}

func (wnd *MainWindow) showMetadata() {
	wnd.MetadataView.SetText(player.StreamTitle)
	var streamLogo = fmt.Sprintf("radio_logos/%s", player.StreamLogo)
	pixbuf, _ := gdk.PixbufNewFromFileAtScale(streamLogo, 32, 32, true)
	wnd.StreamLogoView.SetFromPixbuf(pixbuf)
	go func() {
		metadata, err := player.ShowMetadata()
		if err != nil {
			fmt.Println(err)
			player.Stop()
		}
		title := metadata.Title
		title = strings.TrimSpace(title)
		if title != "" {
			wnd.MetadataView.SetText(title)
		}
	}()
	timerCount++
	fmt.Println(timerCount)
}
