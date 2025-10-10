import re
from google.cloud import texttospeech
from pydub import AudioSegment
import simpleaudio as sa
from io import BytesIO

def speak_natural_text(ssml_to_speak):
    """
    Synthesizes speech (MP3) and plays it with the lightweight simpleaudio library.
    """
    try:
        ssml_clean = re.sub(' +', ' ', ssml_to_speak.replace('\n', '')).strip()
        client = texttospeech.TextToSpeechClient()
        synthesis_input = texttospeech.SynthesisInput(ssml=ssml_clean)
        voice = texttospeech.VoiceSelectionParams(
           language_code="en-US", name="en-US-Wavenet-F"
        )
        audio_config = texttospeech.AudioConfig(
            audio_encoding=texttospeech.AudioEncoding.MP3
        )

        print("Synthesizing speech (MP3)...")
        response = client.synthesize_speech(
            input=synthesis_input, voice=voice, audio_config=audio_config
        )

        audio_segment = AudioSegment.from_file(BytesIO(response.audio_content), format="mp3")

        # Use simpleaudio for playback
        play_obj = sa.play_buffer(
            audio_segment.raw_data,
            num_channels=audio_segment.channels,
            bytes_per_sample=audio_segment.sample_width,
            sample_rate=audio_segment.frame_rate
        )
        # Wait for playback to finish
        play_obj.wait_done()

    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        print("Audio finished.")


if __name__ == "__main__":
    with open('description.txt') as f:
        description = f.read()
        speak_natural_text(description)
