def split_file(input_file, num_lines_per_file):
  """
  Splits a file into chunks of a specified number of lines.

  Args:
      input_file: Path to the input file.
      num_lines_per_file: Number of lines per output file.
  """
  with open(input_file, 'r') as infile:
    lines = infile.readlines()
    lines[-1] = lines[-1].rstrip('\n')
    
    file_num = 1
    current_lines = []
    for line in lines:
      current_lines.append(line)
      if len(current_lines) == num_lines_per_file:
        # Write current chunk to a new file
        output_file = f"{input_file}_part_{file_num}.txt"
        current_lines[-1] = current_lines[-1].rstrip('\n')
        with open(output_file, 'w') as outfile:
          outfile.writelines(current_lines)
        current_lines = []
        file_num += 1

    # Write remaining lines to a separate file (if any)
    if current_lines:
      output_file = f"{input_file}_part_{file_num}.txt"
      with open(output_file, 'w') as outfile:
        outfile.writelines(current_lines)

# Example usage
input_file = "input"
num_lines_per_file = 100
split_file(input_file, num_lines_per_file)

print(f"File '{input_file}' successfully split into chunks of {num_lines_per_file} lines.")
