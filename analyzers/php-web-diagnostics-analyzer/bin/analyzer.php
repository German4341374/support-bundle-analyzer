#!/usr/bin/env php
<?php

declare(strict_types=1);

use SupportBundleAnalyzer\PhpWeb\Analyzer;

require dirname(__DIR__) . '/vendor/autoload.php';

$analyzer = new Analyzer();
while (false !== ($line = fgets(STDIN))) {
    try {
        $request = json_decode($line, true, 32, JSON_THROW_ON_ERROR);
        if (!is_array($request) || '1' !== ($request['protocolVersion'] ?? null)) {
            throw new InvalidArgumentException('Unsupported protocol version.');
        }
        $artifact = $request['artifact'] ?? null;
        $context = $request['context'] ?? [];
        if (!is_array($artifact) || !is_string($artifact['path'] ?? null) || !is_array($context)) {
            throw new InvalidArgumentException('A valid artifact path is required.');
        }
        $root = realpath(is_string($context['workspaceRoot'] ?? null) ? $context['workspaceRoot'] : '.');
        if (false === $root) {
            throw new InvalidArgumentException('Workspace root is unavailable.');
        }
        $target = realpath($root . DIRECTORY_SEPARATOR . $artifact['path']);
        if (false === $target || !str_starts_with($target, $root . DIRECTORY_SEPARATOR)) {
            throw new InvalidArgumentException('Artifact path escapes workspace root.');
        }
        foreach ($analyzer->analyze($target, $artifact['path']) as $finding) {
            echo json_encode(['type' => 'finding', 'finding' => $finding], JSON_THROW_ON_ERROR | JSON_UNESCAPED_SLASHES), PHP_EOL;
        }
    } catch (Throwable $error) {
        echo json_encode(['type' => 'error', 'error' => ['code' => 'PHP_ANALYZER_INPUT_INVALID', 'message' => $error->getMessage()]], JSON_THROW_ON_ERROR), PHP_EOL;
    }
}
