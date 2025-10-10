import google.generativeai as genai
from PIL import Image
import os
import argparse

try:
    genai.configure(api_key=os.environ["GEMINI_API_KEY"])
except KeyError:
    print("Error: GEMINI_API_KEY environment variable not set.")
    exit()

parser = argparse.ArgumentParser(description="A script with adjustable verbosity levels.")

parser.add_argument(
    "-v", "--verbose",
    action="count",
    default=1,
    help="Increase verbosity level. -v for level 2, -vv for level 3."
)

parser.add_argument("--question", type=str, help="A question to append, if any.")
parser.add_argument("-i", "--image", type=str, help="Image to describe.", required=True)

args = parser.parse_args()

try:
    img = Image.open(args.image)
except FileNotFoundError:
    print(f"Error: '{IMAGE}' not found. Make sure the image is in the same folder.")
    exit()

# Load the model 
# We are using the gemini-pro-vision model, which can handle both text and images.

model = genai.GenerativeModel('models/gemini-2.5-pro')

FIELD = "bioinformatics"
PROMPT_BASE = f"You are an assistant to a data scientist, in the field of {FIELD}. Your task is to describe plots, with minimal interpretation, unless explicitely asked otherwise. The goal is to enable accesibility features in data analysis tools. "

#CONTEXT = "Additional context: the plot represents an UMAP embedding of different clusters for cell types."
CONTEXT = "Additional context: the plot is a violin plot for RNA counts for different identities. "


# This is the prompt. Be specific to get the best results.

prompt_v1 = PROMPT_BASE + "Describe this plot in one clear and concise sentence." + CONTEXT
prompt_v2 = PROMPT_BASE + "Describe the key characteristics of the clusters in this plot, focusing on their relative positions, sizes, and separation. Use four or less bullet points for your description. Enclose answer in <speak> tags, and use basic SSML tags to improve generation, but avoid html tags and <break> in particular." + CONTEXT

match args.verbose:
    case 1:
        prompt = prompt_v1
    case 2:
        prompt = prompt_v2
    case _:
        prompt = prompt_v2

if args.question != "":
    prompt += args.question

print(prompt)

# Generate content

response = model.generate_content([prompt, img])
image_description = response.text

print(image_description)

# We'll save this for the next step, text-to-speech
with open("description.txt", "w") as f:
    f.write(image_description)

