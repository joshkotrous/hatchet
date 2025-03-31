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

class BasicSchemaValidator implements ValidatorType {
  validateFormData(formData: any, schema: RJSFSchema = {}): ValidationData<any> {
    const errors: RJSFValidationError[] = [];
    const errorSchema: any = {};
    
    // Basic validation for required fields
    if (schema && schema.required && Array.isArray(schema.required)) {
      for (const field of schema.required) {
        if (formData === undefined || formData === null || !(field in formData)) {
          const error = {
            name: 'required',
            property: `.${field}`,
            message: `${field} is a required property`,
            stack: `.${field} is a required property`,
          };
          errors.push(error);
          if (!errorSchema[field]) errorSchema[field] = {};
          errorSchema[field].__errors = [error.message];
        }
      }
    }
    
    // Check property types if defined in schema
    if (schema && schema.properties && typeof formData === 'object' && formData !== null) {
      for (const key in schema.properties) {
        if (key in formData) {
          const propSchema = (schema.properties as any)[key];
          const value = formData[key];
          
          if (propSchema.type) {
            let typeError = false;
            
            switch (propSchema.type) {
              case 'string':
                typeError = typeof value !== 'string';
                break;
              case 'number':
              case 'integer':
                typeError = typeof value !== 'number';
                break;
              case 'boolean':
                typeError = typeof value !== 'boolean';
                break;
              case 'array':
                typeError = !Array.isArray(value);
                break;
              case 'object':
                typeError = typeof value !== 'object' || value === null || Array.isArray(value);
                break;
            }
            
            if (typeError) {
              const error = {
                name: 'type',
                property: `.${key}`,
                message: `${key} must be a ${propSchema.type}`,
                stack: `.${key} must be a ${propSchema.type}`,
              };
              errors.push(error);
              if (!errorSchema[key]) errorSchema[key] = {};
              errorSchema[key].__errors = errorSchema[key].__errors || [];
              errorSchema[key].__errors.push(error.message);
            }
          }
        }
      }
    }
    
    return { errors, errorSchema };
  }

  toErrorList(errorSchema: any = {}): RJSFValidationError[] {
    const errors: RJSFValidationError[] = [];
    
    const processErrors = (schema: any, path: string = '') => {
      if (schema.__errors) {
        for (const error of schema.__errors) {
          errors.push({
            name: 'validation',
            property: path,
            message: error,
            stack: `${path}: ${error}`,
          });
        }
      }
      
      for (const key in schema) {
        if (key !== '__errors' && typeof schema[key] === 'object') {
          processErrors(schema[key], path ? `${path}.${key}` : `.${key}`);
        }
      }
    };
    
    processErrors(errorSchema);
    return errors;
  }

  isValid(errorSchema: any = {}): boolean {
    const hasErrors = (schema: any): boolean => {
      if (schema.__errors && schema.__errors.length > 0) {
        return true;
      }
      
      for (const key in schema) {
        if (key !== '__errors' && typeof schema[key] === 'object') {
          if (hasErrors(schema[key])) {
            return true;
          }
        }
      }
      
      return false;
    };
    
    return !hasErrors(errorSchema);
  }

  rawValidation(): any {
    return {};
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
          validator={new BasicSchemaValidator()}
          noHtml5Validate={true}
          onChange={(data) => {
            // Transform the data to unwrap the advanced fields
            const formData = { ...data.formData, ...data.formData.advanced };
            delete formData.advanced;
            setInput((prev) =>
              JSON.stringify({
                ...JSON.parse(prev),
                ...formData,
              }),
            );
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