# speech-and-text
google cloudの `speech to text`と `text to speech` に関するリポジトリ



## 起動方法

### 直接起動

cmd/main.go を実行する前に、Go がインストールされている必要があります。

Go のインストール手順については、[Go 公式サイト](https://go.dev/dl/) を参照してください。

Go がインストールされたら、以下のコマンドを使用して cmd/main.go を実行します。

```bash
go run cmd/main.go
```


### コンテナ起動
以下のコマンドを実行します。

```
make compose-build-up
```

またデバッグをサポートしています。

デバッグ利用の際は `port:2345` にアタッチしてください。


## speech to text

Speech to text サービスを呼び出すには、以下の curl コマンドを使用します。

```bash
curl "http://localhost:8080/?bucket_name=your-bucket-name&audio_file_path=your-audio-file.raw&language_codes=id-ID,cmn-Hans-CN,yue-Hant-HK"
```

上記のコマンドを実行する際は、以下のパラメータを適切に置き換えてください。

- `your-bucket-name`: 音声ファイルが保存されている Cloud Storage バケット名
- `your-audio-file.raw`: 変換する音声ファイルのパス (バケット内)
- `language_codes`:  使用する言語コード (カンマ区切り、例: id-ID,cmn-Hans-CN,yue-Hant-HK)
