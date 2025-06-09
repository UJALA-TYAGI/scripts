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
