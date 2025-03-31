import { CodeEditor } from '@/components/v1/ui/code-editor';
import { Loading } from '@/components/v1/ui/loading';

export interface StepRunOutputProps {
  output: string;
  isLoading: boolean;
  errors: string[];
}

export const StepRunOutput: React.FC<StepRunOutputProps> = ({
  output,
  isLoading,
  errors,
}) => {
  if (isLoading) {
    return <Loading />;
  }

  const getDisplayContent = () => {
    if (errors.length > 0) {
      return errors.map((error) => error.split('\\n')).flat();
    }
    
    try {
      return JSON.parse(output);
    } catch (e) {
      console.error('Failed to parse output as JSON:', e);
      return { 
        error: "Invalid JSON format", 
        message: "The output could not be parsed as JSON"
      };
    }
  };

  return (
    <>
      <CodeEditor
        language="json"
        className="mb-4"
        height="400px"
        copy={true}
        code={JSON.stringify(getDisplayContent(), null, 2)}
      />
    </>
  );
};