import LoggingComponent from '@/components/v1/cloud/logging/logs';
import DOMPurify from 'dompurify';

export default function StepRunError({ text }: { text: string }) {
  // Sanitize the text to prevent XSS attacks
  const sanitizedText = DOMPurify.sanitize(text);
  
  return (
    <div className="p-0 w-full min-w-[500px] rounded-md h-[400px] overflow-y-auto">
      <LoggingComponent
        logs={[
          {
            line: sanitizedText,
          },
        ]}
        onTopReached={() => {}}
        onBottomReached={() => {}}
        autoScroll={false}
      />
    </div>
  );
}