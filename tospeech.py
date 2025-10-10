from google.cloud import texttospeech
from pydub import AudioSegment
import simpleaudio as sa
from io import BytesIO

def speak_natural_text(text_to_speak):
    """
    Synthesizes speech using Google Cloud and plays it with simpleaudio.
    """
    try:
        client = texttospeech.TextToSpeechClient()
        synthesis_input = texttospeech.SynthesisInput(text=text_to_speak)
        voice = texttospeech.VoiceSelectionParams(
            language_code="en-US",
            name="en-US-Wavenet-F"
        )
        audio_config = texttospeech.AudioConfig(
            audio_encoding=texttospeech.AudioEncoding.MP3
        )

        print("Synthesizing natural speech...")
        response = client.synthesize_speech(
            input=synthesis_input, voice=voice, audio_config=audio_config
        )

        # 1. Load the MP3 bytes into pydub
        audio_segment = AudioSegment.from_file(BytesIO(response.audio_content), format="mp3")

        print("Playing audio...")
        # 2. Start playback using simpleaudio
        play_obj = sa.play_buffer(
            audio_segment.raw_data,
            num_channels=audio_segment.channels,
            bytes_per_sample=audio_segment.sample_width,
            sample_rate=audio_segment.frame_rate
        )

        # 3. Wait for playback to finish
        play_obj.wait_done()

    except Exception as e:
        print(f"An error occurred: {e}")
        print("Please ensure FFmpeg is installed and accessible in your system's PATH.")
        
    finally:
        print("Audio finished.")


if __name__ == "__main__":
    with open('description.txt') as f:
        description = f.read()
        print(description)
        speak_natural_text(description)
