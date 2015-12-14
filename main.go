package webflake

import (
	"fmt"
	"net/http"
	"time"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/tools/blog/atom"
	"encoding/xml"
    "appengine"
    "appengine/urlfetch"
    "appengine/datastore"
)

func init() {
	  http.HandleFunc("/", feedHandler)
	  http.HandleFunc("/update", updateHandler)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	now		:= time.Now()
	// 更新は 17 - 00 まで
	jst := time.FixedZone( "Asia/Tokyo", 9*60*60 )
    nowJST := now.In( jst )
	if ( nowJST.Hour() < 17 ) {
		fmt.Fprintf( w, "closed (at pm5 will open)" )
		return
	}


	c := appengine.NewContext(r)

	uri := "http://transit.yahoo.co.jp/traininfo/area/4/"

	client := urlfetch.Client(c)
	resp, _ := client.Get(uri)
	doc, _ := goquery.NewDocumentFromResponse(resp)

	person	:= atom.Person { Name: uri }
	var entries []*atom.Entry
	feed	:= atom.Feed {
		Title:		"運行情報 関東 (午後5時から午前0時まで更新)",
		ID:			uri,
		Link:		[]atom.Link { atom.Link{ Rel: "alternate", Href: uri }, },
		Updated:	atom.Time( now ),
		Author:		&person,
	}

	source := ""

	xpath := "#mdStatusTroubleLine .elmTblLstLine tbody tr"

	doc.Find( xpath ).Each( func( _ int, li *goquery.Selection ) {
		s := li.Find( "td:nth-child(1)" )
		if ( len( s.Nodes ) == 0 ) {
			return
		}
		url := ""
		s.Find( "a" ).Each( func( _ int, s *goquery.Selection ) {
			url, _ = s.Attr( "href" )
		} )

		textLine	:= li.Find( "td:nth-child(1)" ).Text()
		textStatus	:= li.Find( "td:nth-child(2)" ).Text()
		textSummary	:= li.Find( "td:nth-child(3)" ).Text()
		e := atom.Entry {
			Title:		"■" + textLine,
			ID:			url,
			Link:		[]atom.Link { { Href: uri } },
			Updated:	atom.Time( now ),
			Content:	&atom.Text { Type: "text", Body: textStatus + ": " + textSummary },
		}
		entries		= append( entries, &e )

		source = source + textLine + textStatus + textSummary
	} )

	// 変更チェック
	var sd StoredData
	keySource := datastore.NewKey( c, "traininfo", "source", 0, nil )
	datastore.Get( c, keySource, &sd )
	// 変更がなければ終了
	if ( string( sd.Serialized ) == source ) {
		fmt.Fprintf( w, "skipped" )
		return
	}
	sd.Serialized	= []byte( source )
	datastore.Put( c, keySource, &sd )

	// 運行情報がなかった場合
	if ( 0 == len( source ) ) {
		e := atom.Entry {
			Title:		"■運行情報",
			ID:			uri,
			Link:		[]atom.Link { { Href: uri } },
			Updated:	atom.Time( now ),
			Content:	&atom.Text { Type: "text", Body: nowJST.Format( "15:04" ) + " 現在、情報はありません" },
		}
		entries		= append( entries, &e )
	}

	// フィードの書き込み
	feed.Entry		= entries

	data, err := xml.Marshal( feed )
	if ( err != nil ) {
		fmt.Println(err)
		return
	}
	sd = StoredData {
		Serialized:  []byte( xml.Header + string( data ) ),
	}

	keyFeed := datastore.NewKey( c, "traininfo", "feed", 0, nil )
	datastore.Put( c, keyFeed, &sd )

	fmt.Fprintf( w, "updated" )
}


func feedHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext( r )

	var sd StoredData

	key := datastore.NewKey( c, "traininfo", "feed", 0, nil )
	datastore.Get( c, key, &sd )

	w.Header().Set( "Content-Type", "application/atom+xml; charset=utf-8" )
	w.WriteHeader( http.StatusOK )
	w.Write( sd.Serialized )
}

