
import sys
import json
import numpy as np
import cv2
import tensorflow as tf

# --- Constants ---
FACENET_MODEL_PATH = "backend/ml/facenet.tflite"
BLAZEFACE_MODEL_PATH = "backend/ml/blazeface.tflite"

# Input image settings
FACENET_INPUT_SIZE = (160, 160)
BLAZEFACE_INPUT_SIZE = (128, 128)

# --- Load TFLite models ---
try:
    facenet_interpreter = tf.lite.Interpreter(model_path=FACENET_MODEL_PATH)
    facenet_interpreter.allocate_tensors()
    facenet_input_details = facenet_interpreter.get_input_details()
    facenet_output_details = facenet_interpreter.get_output_details()

    blazeface_interpreter = tf.lite.Interpreter(model_path=BLAZEFACE_MODEL_PATH)
    blazeface_interpreter.allocate_tensors()
    blazeface_input_details = blazeface_interpreter.get_input_details()
    blazeface_output_details = blazeface_interpreter.get_output_details()
except Exception as e:
    print(json.dumps({"error": f"Failed to load models: {e}"}), file=sys.stderr)
    sys.exit(1)

def preprocess_image(image, target_size):
    """Preprocesses image for TFLite model input."""
    # Resize
    img_resized = cv2.resize(image, target_size)
    # Convert to float32 and normalize
    img_normalized = (np.float32(img_resized) - 127.5) / 128.0
    # Add batch dimension
    return np.expand_dims(img_normalized, axis=0)

def detect_face(image):
    """Detects the most prominent face in an image using BlazeFace."""
    original_h, original_w, _ = image.shape
    
    # Preprocess for BlazeFace
    input_tensor = preprocess_image(image, BLAZEFACE_INPUT_SIZE)
    
    # Run BlazeFace inference
    blazeface_interpreter.set_tensor(blazeface_input_details[0]['index'], input_tensor)
    blazeface_interpreter.invoke()
    
    # Process output: The model outputs detections and scores.
    # Output tensor at index 0 contains bounding boxes, index 1 contains scores.
    detections = blazeface_interpreter.get_tensor(blazeface_output_details[0]['index'])[0]
    scores = blazeface_interpreter.get_tensor(blazeface_output_details[1]['index'])[0]

    if len(scores) == 0 or np.max(scores) < 0.5: # Confidence threshold
        return None, "No face detected with sufficient confidence."

    # Find the detection with the highest score
    best_detection_idx = np.argmax(scores)
    detection = detections[best_detection_idx]

    # Bounding box coordinates are relative to the input size (128x128)
    # and need to be scaled to the original image dimensions.
    # The format is [ymin, xmin, ymax, xmax]
    ymin, xmin, ymax, xmax = detection
    
    x_min = int(xmin * original_w)
    x_max = int(xmax * original_w)
    y_min = int(ymin * original_h)
    y_max = int(ymax * original_h)

    # Ensure coordinates are within image bounds
    x_min = max(0, x_min)
    y_min = max(0, y_min)
    x_max = min(original_w, x_max)
    y_max = min(original_h, y_max)

    # Crop the face from the original image
    cropped_face = image[y_min:y_max, x_min:x_max]
    
    if cropped_face.size == 0:
        return None, "Failed to crop face from image."

    return cropped_face, None

def get_embedding(face_image):
    """Generates a face embedding using the FaceNet model."""
    # Preprocess for FaceNet
    input_tensor = preprocess_image(face_image, FACENET_INPUT_SIZE)
    
    # Run FaceNet inference
    facenet_interpreter.set_tensor(facenet_input_details[0]['index'], input_tensor)
    facenet_interpreter.invoke()
    
    # Get the embedding vector
    embedding = facenet_interpreter.get_tensor(facenet_output_details[0]['index'])[0]
    
    # Normalize the embedding vector (L2 normalization)
    embedding = embedding / np.linalg.norm(embedding)
    
    return embedding.tolist()

def main():
    if len(sys.argv) != 2:
        print(json.dumps({"error": "Image path argument is required."}), file=sys.stderr)
        sys.exit(1)

    image_path = sys.argv[1]
    try:
        image = cv2.imread(image_path)
        if image is None:
            raise IOError(f"Could not read image from path: {image_path}")
        # Convert from BGR (OpenCV default) to RGB
        image_rgb = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
    except Exception as e:
        print(json.dumps({"error": f"Failed to load or process image: {e}"}), file=sys.stderr)
        sys.exit(1)

    # 1. Detect face
    face_image, err_msg = detect_face(image_rgb)
    if err_msg:
        print(json.dumps({"error": err_msg}), file=sys.stderr)
        sys.exit(1)

    # 2. Get embedding
    embedding = get_embedding(face_image)

    # 3. Print embedding as JSON to stdout
    print(json.dumps(embedding))

if __name__ == "__main__":
    main()
