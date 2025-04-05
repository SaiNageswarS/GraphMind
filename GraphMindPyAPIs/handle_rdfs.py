from rdflib import Graph
import os


def unify_ttl(base_folder: str, output_file: str) -> str:
    unified_graph = Graph()

    # Walk through the folder recursively
    for root, dirs, files in os.walk(base_folder):
        for file in files:
            if file.endswith(".ttl"):
                file_path = os.path.join(root, file)
                try:
                    # Parse the Turtle file and merge into the unified graph
                    unified_graph.parse(file_path, format="turtle")
                    print(f"Parsed {file_path}")
                except Exception as e:
                    print(f"Failed to parse {file_path}: {e}")

    # Serialize the unified graph to the output file in Turtle format
    unified_graph.serialize(destination=output_file, format="turtle")
    print(f"Unified graph saved to {output_file}")
    return output_file

