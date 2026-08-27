<?php

declare(strict_types=1);

$finder = PhpCsFixer\Finder::create()->in([__DIR__.'/src', __DIR__.'/tests', __DIR__.'/bin']);

return (new PhpCsFixer\Config())
    ->setRiskyAllowed(true)
    ->setRules(['@PER-CS2.0' => true, 'declare_strict_types' => true, 'strict_param' => true])
    ->setFinder($finder);
