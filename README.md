 webflake
==========

# はじめに
定期的に [Yahoo! の 路線情報](http://transit.yahoo.co.jp/traininfo/area/4/) を取得して Atom フィードにします。

# 動作環境
Google App Engine を想定しています。
17-24 時の間だけデータを更新します。また、前回と同じ情報だった場合はデータ更新をスキップします。

# API
デプロイ先のルートアクセスでフィードを返します。
/update で更新をします。

# デプロイ方法
アプリケーションキーを取得して app.yaml を書いてください。その後、
```
goapp deploy
```
でデプロイされるはずです。一時間ごとに cron で取得に行きます。
