import api, { StepRun, StepRunStatus, WorkflowRun, queries } from '@/lib/api';
import { useEffect, useMemo, useState } from 'react';
import { RunStatus } from '../../components/run-statuses';
import { Button } from '@/components/ui/button';
import invariant from 'tiny-invariant';
import { useApiError } from '@/lib/hooks';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { cn } from '@/lib/utils';
import { useOutletContext } from 'react-router-dom';
import { TenantContextType } from '@/lib/outlet';
import { PlayIcon } from '@radix-ui/react-icons';
import { StepRunOutput } from './step-run-output';
import { StepRunInputs } from './step-run-inputs';
import { Loading } from '@/components/ui/loading';
import { StepStatusDetails } from '..';
import {
  TooltipProvider,
  Tooltip,
  TooltipTrigger,
  TooltipContent,
} from '@/components/ui/tooltip';
import { VscNote, VscJson } from 'react-icons/vsc';

export function StepRunPlayground({
  stepRun,
  setStepRun,
  workflowRun,
}: {
  stepRun: StepRun | undefined;
  setStepRun: (stepRun: StepRun | undefined) => void;
  workflowRun: WorkflowRun;
}) {
  const { tenant } = useOutletContext<TenantContextType>();
  invariant(tenant);

  const [errors, setErrors] = useState<string[]>([]);

  const { handleApiError } = useApiError({
    setErrors,
  });

  const updateParentData = (input: any, workflowRun: WorkflowRun) => {
    // HACK this is a temporary solution to get the parent data from the previous run
    // this should be handled by the backend
    const parents = Object.keys(input.parents);
    if (!workflowRun.jobRuns || !workflowRun.jobRuns[0]) {
      return input;
    }

    return workflowRun.jobRuns[0].stepRuns?.reduce((acc, stepRun) => {
      if (!stepRun.step || !parents.includes(stepRun.step.readableId)) {
        return acc;
      }

      return {
        ...acc,
        [stepRun.step.readableId]: JSON.parse(stepRun.output || '{}'),
      };
    }, {});
  };

  const originalInput = useMemo(() => {
    const input = JSON.parse(stepRun?.input || '{}');
    const previousRunData = updateParentData(input, workflowRun);

    const modifiedInput = JSON.stringify(
      {
        ...input,
        parents: {
          ...input.parents,
          ...previousRunData,
        },
      },
      null,
      2,
    );

    return modifiedInput;
  }, [stepRun?.input, workflowRun]);

  const [stepInput, setStepInput] = useState(originalInput);

  useEffect(() => {
    setStepInput(originalInput);
  }, [originalInput]);

  const getStepRunQuery = useQuery({
    ...queries.stepRuns.get(tenant.metadata.id, stepRun?.metadata.id || ''),
    enabled: !!stepRun,
    refetchInterval: (query) => {
      const data = query.state.data;

      if (
        data?.status != 'SUCCEEDED' &&
        data?.status != 'FAILED' &&
        data?.status != 'CANCELLED'
      ) {
        return 1000;
      }
    },
  });

  const queryClient = useQueryClient();

  const rerunStepMutation = useMutation({
    mutationKey: [
      'step-run:update:rerun',
      stepRun?.tenantId,
      stepRun?.metadata.id,
    ],
    mutationFn: async (input: object) => {
      invariant(stepRun?.tenantId, 'has tenantId');

      const res = await api.stepRunUpdateRerun(
        stepRun?.tenantId,
        stepRun?.metadata.id,
        {
          input: input,
        },
      );

      return res.data;
    },
    onMutate: () => {
      setErrors([]);
    },
    onSuccess: (stepRun: StepRun) => {
      queryClient.invalidateQueries({
        queryKey: queries.workflowRuns.get(
          tenant.metadata.id,
          workflowRun.metadata.id,
        ).queryKey,
      });

      setStepRun(stepRun);
      getStepRunQuery.refetch();
    },
    onError: handleApiError,
  });

  useEffect(() => {
    if (getStepRunQuery.data) {
      setStepRun(getStepRunQuery.data);
    }
  }, [getStepRunQuery.data, setStepRun]);

  const output = stepRun?.output || '{}';

  const COMPLETED = ['SUCCEEDED', 'FAILED', 'CANCELLED'];
  const isLoading = !COMPLETED.includes(stepRun?.status || '');

  const handleOnPlay = () => {
    const inputObj = JSON.parse(stepInput);
    rerunStepMutation.mutate(inputObj);
  };

  const [mode, setMode] = useState<'form' | 'json'>(
    (localStorage.getItem('mode') as 'form' | 'json') || 'form',
  );

  useEffect(() => {
    localStorage.setItem('mode', mode);
  }, [mode]);

  const handleModeSwitch = () => {
    setMode((prev) => (prev === 'json' ? 'form' : 'json'));
  };

  const disabled = rerunStepMutation.isPending || isLoading;

  // Function to detect the operating system
  const getOS = () => {
    const userAgent = window.navigator.userAgent;
    // Simple checks for platform; these could be extended as needed
    if (userAgent.includes('Mac')) {
      return 'MacOS';
    } else if (userAgent.includes('Win')) {
      return 'Windows';
    } else {
      // Default or other OS
      return 'unknown';
    }
  };
  // Determine the appropriate shortcut based on the OS
  const shortcut = getOS() === 'MacOS' ? 'Cmd + Enter' : 'Ctrl + Enter';

  return (
    <div className="">
      {stepRun && (
        <>
          <div className="flex flex-row gap-2 justify-between items-center sticky top-0 z-50">
            <div className="text-2xl font-semibold tracking-tight">
              Playground/{stepRun?.step?.readableId}
            </div>
            <div className="flex flex-row gap-2 justify-end items-center">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={handleModeSwitch}
                    >
                      {mode === 'json' && <VscNote className="h-4 w-4" />}
                      {mode === 'form' && <VscJson className="h-4 w-4" />}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    {mode === 'json' && 'Switch to Form Mode'}
                    {mode === 'form' && 'Switch to JSON Mode'}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      className="w-fit"
                      disabled={disabled}
                      onClick={handleOnPlay}
                    >
                      {disabled ? (
                        <>
                          <Loading />
                          Playing
                        </>
                      ) : (
                        <>
                          <PlayIcon className={cn('h-4 w-4 mr-2')} />
                          Replay Step
                        </>
                      )}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>{shortcut}</TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </div>
          <div className="flex flex-row gap-4 mt-4">
            <div className="flex-grow w-1/2">
              Inputs
              {stepInput && (
                <StepRunInputs
                  schema={stepRun.inputSchema || ''}
                  input={stepInput}
                  setInput={setStepInput}
                  disabled={disabled}
                  handleOnPlay={handleOnPlay}
                  mode={mode}
                />
              )}
            </div>
            <div className="flex-grow flex-col flex gap-4 w-1/2 ">
              <div className="flex flex-col">
                <div className="flex flex-row justify-between items-center mb-4">
                  <div>Outputs</div>
                  <RunStatus
                    status={
                      errors.length > 0
                        ? StepRunStatus.FAILED
                        : stepRun?.status || StepRunStatus.PENDING
                    }
                  />
                </div>
                <StepRunOutput
                  stepRun={stepRun}
                  output={output}
                  isLoading={isLoading}
                  errors={
                    [
                      ...errors,
                      stepRun.error
                        ? StepStatusDetails({ stepRun })
                        : undefined,
                    ].filter((e) => !!e) as string[]
                  }
                />
              </div>
            </div>
          </div>
        </>
      )}
      {errors.length > 0 && <div className="mt-4"></div>}
    </div>
  );
}
