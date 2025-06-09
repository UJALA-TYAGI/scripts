index.js
import React from "react";
import ReactDOM from "react-dom/client";
import SurveyUI from "./SurveyUI";
import "./index.css";

const root = ReactDOM.createRoot(document.getElementById("root"));
root.render(
  <React.StrictMode>
    <SurveyUI />
  </React.StrictMode>
);

index.css
@tailwind base;
@tailwind components;
@tailwind utilities;

surveyForm.jsx
import React from "react";

const SurveyForm = ({ surveyName, questions, onChange, onSubmit, error }) => {
  return (
    <div className="min-h-screen bg-gradient-to-tr from-green-50 to-white flex items-center justify-center px-4 py-12">
      <div className="bg-white shadow-xl rounded-2xl p-10 max-w-3xl w-full border border-gray-200">
        <h1 className="text-3xl font-bold mb-8 text-center text-emerald-700">{surveyName}</h1>

        {error && (
          <div className="mb-4 text-sm text-red-500 text-center">
            Could not connect to backend, using mock data instead.
          </div>
        )}

        <form onSubmit={onSubmit} className="space-y-10">
          {questions.map((q, idx) => (
            <div key={q.id} className="space-y-2">
              <label className="block text-lg font-medium text-gray-800">
                {idx + 1}. {q.question}
              </label>

              {q.type === "mcq" ? (
                <div className="flex flex-col space-y-2 pl-4">
                  {q.options.map((opt) => (
                    <label key={opt} className="inline-flex items-center text-gray-700">
                      <input
                        type="radio"
                        name={q.id}
                        value={opt}
                        onChange={() => onChange(q.id, opt)}
                        className="accent-emerald-600 mr-2"
                        required
                      />
                      {opt}
                    </label>
                  ))}
                </div>
              ) : (
                <textarea
                  className="w-full mt-2 p-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                  rows="4"
                  placeholder="Your answer..."
                  onChange={(e) => onChange(q.id, e.target.value)}
                  required
                />
              )}
            </div>
          ))}

          <button
            type="submit"
            className="w-full mt-4 bg-emerald-600 text-white py-3 px-6 rounded-lg text-lg font-semibold hover:bg-emerald-700 transition duration-200"
          >
            Submit Feedback
          </button>
        </form>
      </div>
    </div>
  );
};

export default SurveyForm;


SurveyUI.jsx
import React, { useEffect, useState } from "react";
import axios from "axios";
import SurveyForm from "./SurveyForm";

const SurveyUI = () => {
  const [surveyName, setSurveyName] = useState("Loading Survey...");
  const [questions, setQuestions] = useState([]);
  const [responses, setResponses] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const mockSurvey = {
    surveyName: "Event Feedback Survey (Fallback)",
    questions: [
      {
        id: "q1",
        question: "How would you rate the event?",
        type: "mcq",
        options: ["Excellent", "Good", "Average", "Poor"],
      },
      {
        id: "q2",
        question: "What did you like most about the event?",
        type: "text",
      },
    ],
  };

  useEffect(() => {
    axios
      .get("http://localhost:5000/api/survey")
      .then((res) => {
        setSurveyName(res.data.surveyName);
        setQuestions(res.data.questions);
      })
      .catch(() => {
        setSurveyName(mockSurvey.surveyName);
        setQuestions(mockSurvey.questions);
        setError(true);
      })
      .finally(() => setLoading(false));
  }, []);

  const handleChange = (qid, value) => {
    setResponses((prev) => ({ ...prev, [qid]: value }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    if (error) {
      // Save to file locally
      const fileData = JSON.stringify(responses, null, 2);
      const blob = new Blob([fileData], { type: "application/json" });
      const url = URL.createObjectURL(blob);

      const link = document.createElement("a");
      link.download = `${surveyName.replace(/\s+/g, "_")}_responses.json`;
      link.href = url;
      link.click();

      alert("Mock survey responses saved as a JSON file!");
    } else {
      // Send to backend
      fetch("http://localhost:5000/api/submit", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(responses),
      })
        .then((res) => {
          if (res.ok) {
            alert("Responses submitted successfully!");
          } else {
            alert("Submission failed. Please try again.");
          }
        })
        .catch(() => {
          alert("Network error while submitting responses.");
        });
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-100">
        <p className="text-xl text-gray-700 animate-pulse">Loading survey...</p>
      </div>
    );
  }

  return (
    <SurveyForm
      surveyName={surveyName}
      questions={questions}
      onChange={handleChange}
      onSubmit={handleSubmit}
      error={error}
    />
  );
};

export default SurveyUI;

tailwincss
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {},
  },
  plugins: [],
};




useEffect(() => {
  axios
    .get("http://localhost:5000/expose_data")
    .then((res) => {
      const rawData = res.data.data;

      const parsedQuestions = rawData.map((q) => {
        let options = [];

        try {
          if (q.options) {
            // Replace single quotes with double and parse
            options = JSON.parse(q.options.replace(/'/g, '"'));
          }
        } catch (err) {
          console.warn("Failed to parse options for question:", q.id);
        }

        return {
          ...q,
          id: q.id || Math.random(),
          type: q.type?.trim(),
          options,
          scale_min: isNaN(q.scale_min) ? null : q.scale_min,
          scale_max: isNaN(q.scale_max) ? null : q.scale_max,
        };
      });

      setSurveyName("Event Feedback Survey");
      setQuestions(parsedQuestions);
    })
    .catch((err) => {
      console.error("Error fetching survey:", err);
      setSurveyName(mockSurvey.surveyName);
      setQuestions(mockSurvey.questions);
      setError(true);
    })
    .finally(() => setLoading(false));
}, []);



import React from "react";

const SurveyForm = ({ surveyName, questions, onChange, onSubmit, error }) => {
  return (
    <form onSubmit={onSubmit} className="max-w-3xl mx-auto p-6 bg-white shadow rounded">
      <h1 className="text-2xl font-bold mb-6 text-center">{surveyName}</h1>

      {questions.map((q) => (
        <div key={q.id} className="mb-6">
          <label className="block text-lg font-medium mb-2">{q.question}</label>

          {q.type === "single_choice" &&
            q.options?.map((opt, idx) => (
              <label key={idx} className="flex items-center mb-1">
                <input
                  type="radio"
                  name={`q${q.id}`}
                  value={opt}
                  onChange={(e) => onChange(q.id, e.target.value)}
                  className="accent-blue-500"
                />
                <span className="ml-2">{opt}</span>
              </label>
            ))}

          {q.type === "multiple_choice" &&
            q.options?.map((opt, idx) => (
              <label key={idx} className="flex items-center mb-1">
                <input
                  type="checkbox"
                  value={opt}
                  onChange={(e) =>
                    onChange(q.id, (prev = []) =>
                      e.target.checked ? [...prev, opt] : prev.filter((o) => o !== opt)
                    )
                  }
                  className="accent-green-500"
                />
                <span className="ml-2">{opt}</span>
              </label>
            ))}

          {q.type === "text" && (
            <textarea
              rows={3}
              className="w-full p-2 border rounded"
              onChange={(e) => onChange(q.id, e.target.value)}
            />
          )}

          {q.type === "scale" && q.scale_min != null && q.scale_max != null && (
            <div className="flex items-center space-x-4">
              <span>{q.scale_min}</span>
              <input
                type="range"
                min={q.scale_min}
                max={q.scale_max}
                onChange={(e) => onChange(q.id, Number(e.target.value))}
                className="w-full"
              />
              <span>{q.scale_max}</span>
            </div>
          )}
        </div>
      ))}

      <button
        type="submit"
        className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700"
      >
        Submit
      </button>
    </form>
  );
};

export default SurveyForm;



from punetacone import jsonify, APIBlueprint as Blueprint, Tag
import pandas as pd
import os
import numpy as np
from . import logger  # Make sure logger is defined or imported

app = Blueprint("app", __name__)
api_tag = Tag(name="Pyneta Base", description="Welcome API for users.")

@app.post("/expose_data", summary="Expose Data", tags=[api_tag])
def expose_data():
    """
    Expose Data Endpoint - POST Request
    Reads from data.xlsx and returns structured survey data
    """
    excel_path = os.path.join(os.path.dirname(__file__), "data.xlsx")
    if not os.path.exists(excel_path):
        return jsonify({"status": "error", "message": "data.xlsx not found"}), 404

    df = pd.read_excel(excel_path)

    # Define the clean function here
    def clean(row):
        options = []
        if pd.notna(row.get("options")):
            try:
                # Convert string to list (e.g., "['A', 'B']" → ['A', 'B'])
                options = eval(row["options"]) if isinstance(row["options"], str) else row["options"]
            except Exception as e:
                logger.warning(f"Could not parse options: {e}")
                options = []

        return {
            "id": int(row["id"]),
            "question": row["question"],
            "type": row["type"].strip().lower(),
            "options": options,
            "scale_min": int(row["scale_min"]) if not pd.isna(row["scale_min"]) else None,
            "scale_max": int(row["scale_max"]) if not pd.isna(row["scale_max"]) else None,
        }

    # Apply cleaning
    cleaned_data = [clean(row) for _, row in df.iterrows()]

    logger.info(f"Returning cleaned survey data with {len(cleaned_data)} questions.")
    return jsonify({"status": "success", "data": cleaned_data})


