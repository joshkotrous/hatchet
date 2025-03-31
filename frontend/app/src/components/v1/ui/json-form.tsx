import { cn } from '@/lib/utils';
import {
  RJSFSchema,
  RJSFValidationError,
  UiSchema,
  ValidationData,
  ValidatorType,
} from '@rjsf/utils';
import Form from '@rjsf/core';
import { PlayIcon } from '@radix-ui/react-icons';
import { Button } from './button';
import { Loading } from './loading';
import { CollapsibleSection } from './form-inputs/collapsible-section';
import { DynamicSizeInputTemplate } from './form-inputs/dynamic-size-input-template';
import { createContext, useRef } from 'react';

type JSONPrimitive = string | number | boolean | null | Array<JSONPrimitive>;
export type JSONType = {
  [key: string]: JSONType | JSONPrimitive | Array<JSONType>;
};

export const DEFAULT_COLLAPSED = ['advanced', 'user data'];

class BasicFormValidator implements ValidatorType {
  validateFormData(formData: any, schema: RJSFSchema = {}): ValidationData<any> {
    const errors: RJSFValidationError[] = [];
    const errorSchema: Record<string, any> = {};
    
    // Check required fields
    if (schema.required && Array.isArray(schema.required) && typeof formData === 'object' && formData !== null) {
      for (const field of schema.required) {
        if (formData[field] === undefined) {
          const error: RJSFValidationError = {
            property: field,
            message: `${field} is a required field`,
            stack: `${field} is a required field`
          };
          errors.push(error);
          
          errorSchema[field] = errorSchema[field] || {};
          errorSchema[field].__errors = errorSchema[field].__errors || [];
          errorSchema[field].__errors.push(error.message);
        }
      }
    }
    
    // Basic type validation
    if (schema.properties && typeof formData === 'object' && formData !== null) {
      for (const [field, fieldSchema] of Object.entries(schema.properties)) {
        if (formData[field] !== undefined && typeof fieldSchema === 'object' && 'type' in fieldSchema) {
          const expectedType = (fieldSchema as RJSFSchema).type;
          let isValid = true;
          
          switch (expectedType) {
            case 'string':
              isValid = typeof formData[field] === 'string';
              break;
            case 'number':
            case 'integer':
              isValid = typeof formData[field] === 'number';
              break;
            case 'boolean':
              isValid = typeof formData[field] === 'boolean';
              break;
            case 'array':
              isValid = Array.isArray(formData[field]);
              break;
            case 'object':
              isValid = typeof formData[field] === 'object' && formData[field] !== null && !Array.isArray(formData[field]);
              break;
          }
          
          if (!isValid) {
            const error: RJSFValidationError = {
              property: field,
              message: `${field} must be of type ${expectedType}`,
              stack: `${field} must be of type ${expectedType}`
            };
            errors.push(error);
            
            errorSchema[field] = errorSchema[field] || {};
            errorSchema[field].__errors = errorSchema[field].__errors || [];
            errorSchema[field].__errors.push(error.message);
          }
        }
      }
    }
    
    return { errors, errorSchema };
  }

  toErrorList(errorSchema: Record<string, any> = {}): RJSFValidationError[] {
    const errors: RJSFValidationError[] = [];
    
    for (const key in errorSchema) {
      if (key === '__errors' && Array.isArray(errorSchema[key])) {
        for (const message of errorSchema[key]) {
          errors.push({
            property: '',
            message,
            stack: message
          });
        }
      } else if (typeof errorSchema[key] === 'object') {
        const childErrors = this.toErrorList(errorSchema[key]);
        for (const error of childErrors) {
          errors.push({
            property: key + (error.property ? '.' + error.property : ''),
            message: error.message,
            stack: key + (error.stack ? '.' + error.stack : '')
          });
        }
      }
    }
    
    return errors;
  }

  isValid(formData: any, schema: RJSFSchema): boolean {
    const { errors } = this.validateFormData(formData, schema);
    return errors.length === 0;
  }

  rawValidation(formData: any, schema: RJSFSchema): ValidationData<any> {
    return this.validateFormData(formData, schema);
  }
}

interface JSONFormContextSchema {
  form?: React.RefObject<Form>;
}

export const JSONFormContext = createContext<JSONFormContextSchema>({
  form: undefined,
});

export function JsonForm({
  inputSchema,
  inputData,
  className,
  setInput,
  disabled,
  onSubmit,
}: {
  inputSchema: JSONType;
  className?: string;
  inputData: JSONType;
  setInput: React.Dispatch<React.SetStateAction<string>>;
  disabled?: boolean;
  onSubmit: () => void;
}) {
  const formRef = useRef<Form>(null);

  const schema = {
    ...inputSchema,
    required: undefined,
    $schema: undefined,
    properties: {
      ...(inputSchema.properties as any),
      triggered_by: undefined,
      advanced: {
        // Transform the schema to wrap the triggered by field
        type: 'object',
        properties: {
          triggered_by: inputSchema.properties
            ? (inputSchema.properties as any).triggered_by
            : undefined,
        },
      },
    },
  } as RJSFSchema;

  delete schema.properties?.triggered_by;

  const uiSchema: UiSchema<any, RJSFSchema, any> = {
    input: {
      'ui:title': 'workflow input',
    },
    parents: {
      'ui:title': 'parent step data',
    },
    overrides: {
      'ui:title': 'step overrides',
    },
    user_data: {
      'ui:title': 'user data',
    },
    'ui:order': ['input', 'overrides', 'parents', '*'],
  };

  return (
    <JSONFormContext.Provider value={{ form: formRef }}>
      <div
        className={cn(
          className,
          'w-full h-fit relative rounded-lg overflow-hidden',
        )}
      >
        <Form
          ref={formRef}
          formData={inputData}
          schema={schema}
          disabled={disabled}
          templates={{
            BaseInputTemplate: DynamicSizeInputTemplate,
            ObjectFieldTemplate: CollapsibleSection,
          }}
          uiSchema={uiSchema}
          validator={new BasicFormValidator()}
          noHtml5Validate={true}
          onChange={(data) => {
            // Create a clean copy of the form data
            const formData: JSONType = { ...data.formData };
            
            // Safely extract only the expected triggered_by field from advanced
            if (formData.advanced && typeof formData.advanced === 'object') {
              // Extract only the triggered_by property from advanced
              const advanced = formData.advanced as JSONType;
              if ('triggered_by' in advanced) {
                formData.triggered_by = advanced.triggered_by;
              }
            }
            
            // Remove the advanced object to prevent unwanted fields
            delete formData.advanced;
            
            // Update state with validated data
            setInput((prev) => {
              try {
                const prevData = JSON.parse(prev) as JSONType;
                return JSON.stringify({
                  ...prevData,
                  ...formData,
                });
              } catch (error) {
                console.error('Error parsing previous form data:', error);
                return JSON.stringify(formData);
              }
            });
          }}
          onSubmit={onSubmit}
          onError={(e) => {
            console.error(e);
          }}
        >
          <Button className="w-fit invisible" disabled={disabled}>
            {disabled ? (
              <>
                <Loading />
                Playing
              </>
            ) : (
              <>
                <PlayIcon
                  className={cn(disabled ? 'rotate-180' : '', 'h-4 w-4 mr-2')}
                />
                Play Step
              </>
            )}
          </Button>
        </Form>
      </div>
    </JSONFormContext.Provider>
  );
}