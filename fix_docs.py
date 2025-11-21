import re
import os

file_path = 'website/src/docs/proposal-2-context-based-cookbook.md'

if not os.path.exists(file_path):
    print(f"File not found: {file_path}")
    exit(1)

with open(file_path, 'r') as f:
    content = f.read()

# Regex to find code blocks
# We assume code blocks start with ``` and end with ``` on their own lines.
# We capture the whole block.
pattern = re.compile(r'(^```[\w-]*\n.*?^```)', re.MULTILINE | re.DOTALL)

def replace_block(match):
    block = match.group(1)
    if '{{' in block:
        return f"::: v-pre\n{block}\n:::"
    return block

new_content = pattern.sub(replace_block, content)

with open(file_path, 'w') as f:
    f.write(new_content)
print("Fixed docs.")
