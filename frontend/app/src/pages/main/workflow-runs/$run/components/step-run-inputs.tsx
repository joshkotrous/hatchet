import { CodeEditor } from '@/components/ui/code-editor';
import { JSONType, JsonForm } from '@/components/ui/json-form';
import { useEffect, useState, useMemo } from 'react';

export interface StepRunOutputProps {
  input: string;
  schema: object;
  setInput: React.Dispatch<React.SetStateAction<string>>;
  disabled: boolean;
  handleOnPlay: () => void;
  mode: 'json' | 'form';
}

const tryFormat = (input: string) => {
  try {
    return JSON.stringify(JSON.parse(input), null, 2);
  } catch (e) {
    return input;
  }
};

// Safely parse and validate input
const safelyParseInput = (input: string): { data: any; error: string | null } => {
  try {
    const parsed = JSON.parse(input);
    
    // Basic validation to ensure it's an object or array
    if (typeof parsed !== 'object' || parsed === null) {
      return { 
        data: {}, 
        error: "Input must be a valid JSON object or array" 
      };
    }
    
    return { data: parsed, error: null };
  } catch (e) {
    return { 
      data: {}, 
      error: e instanceof Error ? e.message : "Invalid JSON input" 
    };
  }
};

export const StepRunInputs: React.FC<StepRunOutputProps> = ({
  input,
  schema,
  disabled,
  handleOnPlay,
  setInput,
  mode,
}) => {
  const [currentInput, setCurrentInput] = useState(tryFormat(input));
  const [parseError, setParseError] = useState<string | null>(null);

  // Memoize the parsed input to avoid re-parsing on every render
  const parsedInput = useMemo(() => {
    return safelyParseInput(input);
  }, [input]);

  // Update parse error when input changes
  useEffect(() => {
    setParseError(parsedInput.error);
  }, [parsedInput]);

  useEffect(() => {
    setCurrentInput(input);
  }, [input]);

  const handleCodeChange = (code: string | undefined) => {
    if (!code) {
      return;
    }
    setCurrentInput(code);
    setInput(code);
  };

  // Handle form submission with validation
  const handleFormSubmit = () => {
    if (!parseError) {
      handleOnPlay();
    }
  };

  return (
    <>
      {mode === 'form' && (
        <div>
          {!schema ? (
            <>No Schema</>
          ) : (
            <>
              {parseError && (
                <div className="text-red-500 mb-2">Error: {parseError}</div>
              )}
              <JsonForm
                inputSchema={schema as JSONType}
                setInput={setInput}
                inputData={parsedInput.data}
                onSubmit={handleFormSubmit}
                disabled={disabled || parseError !== null}
              />
            </>
          )}
        </div>
      )}

      {mode === 'json' && (
        <div>
          <CodeEditor
            language="json"
            className="my-4"
            height="400px"
            code={currentInput}
            setCode={handleCodeChange}
          />
        </div>
      )}
    </>
  );
};