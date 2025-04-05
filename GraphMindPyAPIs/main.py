from flask import Flask, request, jsonify
import os

from handle_rdfs import unify_ttl

app = Flask(__name__)


@app.route('/unify', methods=['POST'])
def unify():
    data = request.get_json()
    if not data or "folder" not in data or "output" not in data:
        return jsonify({"error": "Request must contain 'folder' and 'output' keys."}), 400

    folder = data["folder"]
    output_file = data["output"]

    if not os.path.isdir(folder):
        return jsonify({"error": f"Folder '{folder}' does not exist."}), 400

    try:
        result_path = unify_ttl(folder, output_file)
        return jsonify({"combined_graph_path": result_path}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500


if __name__ == '__main__':
    app.run()
