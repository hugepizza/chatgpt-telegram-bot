# ChatGPT Telegram bot

Run your own ChatGPT Telegram bot!

Support voice message and return voice message (by google/azure tts).

chatGPT的Telegram机器人支持文字输入返回文字，语音输入返回语音，目前有调用google和azure的实现，需要ffmpeg用于将tg的oga音频转为mp3。


## Setup

1. Get your OpenAI API key

   You can create an account on the OpenAI website and [generate your API key](https://platform.openai.com/account/api-keys).

2. Get your telegram bot token

   Create a bot from Telegram [@BotFather](https://t.me/BotFather) and obtain an access token.

3. Get your google tts credentials file

4. Install using ffmpeg

5. Set the environment variables and run

```bash
export OPENAI_API_KEY=<your_openai_api_key>
export TELEGRAM_APITOKEN=<your_telegram_bot_token>

export ALLOWED_TELEGRAM_ID=<your_telegram_id>,<your_friend_telegram_id>

export GOOGLE_TTS_Cert_File=<your_google_tts_cert_file_path>

# if you are going to use azure tts
# export Azure_TTS_Key=<azure_tts_key> 
# export Azure_TTS_Region=<azure_tts_region>

chatgpt-telegram-bot
```
